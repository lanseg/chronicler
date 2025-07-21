package resolver

import (
	"context"
	"fmt"
	"net/url"
	"path/filepath"
	"slices"
	"sync"
	"time"

	"chronicler/adapter"
	cm "chronicler/common"
	opb "chronicler/proto"
	"chronicler/storage"
)

const (
	objectFileName = "snapshot"
)

type resolverTask struct {
	link    *opb.Link
	adapter int
}

type Resolver interface {
	Resolve(link *opb.Link) error
	Start()
	Stop()
	Wait()
}

type resolver struct {
	Resolver

	done       func()
	statusMux  sync.Mutex
	taskWaiter sync.WaitGroup
	tasks      chan resolverTask

	root     string
	loader   cm.Downloader
	adapters []adapter.Adapter
	logger   *cm.Logger
}

func NewResolver(root string, loader cm.Downloader, adapters []adapter.Adapter) Resolver {
	r := &resolver{
		taskWaiter: sync.WaitGroup{},
		tasks:      make(chan resolverTask, 10),
		adapters:   adapters,
		loader:     loader,
		root:       root,
		logger:     cm.NewLogger("Resolver"),
	}
	r.logger.Infof("Initialized resolver with %d adapters", len(adapters))
	return r
}

func (r *resolver) Start() {
	r.statusMux.Lock()
	defer r.statusMux.Unlock()
	if r.done != nil {
		r.logger.Infof("Already started")
		return
	}
	r.logger.Infof("Starting resolver thread")
	ctx, done := context.WithCancel(context.Background())
	r.done = done

	go func() {
	loop:
		for {
			select {
			case <-ctx.Done():
				break loop
			case task := <-r.tasks:
				if err := r.resolveTask(task); err != nil {
					r.logger.Warningf("Cannot resolve link %s: %s", task.link.Href, err)
				}
				r.taskWaiter.Done()
			}
		}
		r.taskWaiter.Wait()
		close(r.tasks)
	}()
}

func (r *resolver) Wait() {
	r.logger.Infof("Waiting for all tasks to complete")
	r.taskWaiter.Wait()
}

func (r *resolver) Stop() {
	r.statusMux.Lock()
	defer r.statusMux.Unlock()
	if r.done == nil {
		r.logger.Infof("Not running")
		return
	}
	r.logger.Infof("Stopping resolver")
	r.done()
}

func (r *resolver) Resolve(link *opb.Link) error {
	for i, adapter := range r.adapters {
		if adapter.Match(link) {
			r.taskWaiter.Add(1)
			r.tasks <- resolverTask{link: link, adapter: i}
			break
		}
	}
	return nil
}

func (r *resolver) getStorage(link *opb.Link) (*storage.BlockStorage, error) {
	ls, err := storage.NewLocalStorage(filepath.Join(r.root, cm.UUID4For(link)))
	if err != nil {
		return nil, err
	}
	return &storage.BlockStorage{
		Storage: ls,
	}, nil
}

func (r *resolver) resolveTask(task resolverTask) error {
	ad := r.adapters[task.adapter]
	link := task.link

	objs, err := ad.Get(link)
	if err != nil {
		return fmt.Errorf("adapter %q cannot get data from %q: %w", ad, link, err)
	}
	s, err := r.getStorage(link)
	if err != nil {
		return fmt.Errorf("cannot open storage for link %q: %w", link, err)
	}

	snapshot := &opb.Snapshot{
		FetchTime: &opb.Timestamp{
			Seconds: time.Now().Unix(),
		},
		Link:    link,
		Objects: objs,
	}

	// Filename could be arbitrary, as the storage sanitizes it if necessary.
	bytesWritten, err := s.PutJSON(&storage.PutRequest{
		Url:             fmt.Sprintf("%s.json", objectFileName),
		SaveOnOverwrite: true,
	}, snapshot)
	if err != nil {
		return fmt.Errorf("cannot save link data for link %q: %w", link, err)
	}
	r.logger.Infof("Saved %q, written bytes: %d", objectFileName, bytesWritten)

	filesToLoad := map[*url.URL]bool{}

	for attachment := range cm.FlatMap(slices.Values(objs), opb.Attachments) {
		if attachment.Mime == "" {
			continue
		}
		fileUrl, err := url.Parse(attachment.Url.Href)
		if err != nil {
			r.logger.Warningf("Cannot parse url %q: %s", fileUrl, err)
			continue
		}
		filesToLoad[fileUrl] = true
	}

	file := 0
	toLoad := len(filesToLoad)
	r.logger.Infof("Files to download: %d", toLoad)
	for k := range filesToLoad {
		file += 1
		r.logger.Infof("Downloading [%d of %d] %s", file, toLoad, k)
		writer, err := s.Put(&storage.PutRequest{Url: k.String()})
		if err != nil {
			r.logger.Warningf("Cannot create writer for %q: %s", k.String(), err)
			continue
		}
		_, err = r.loader.Download(k.String(), writer)
		if err != nil {
			r.logger.Warningf("Failed to download %s: %s", k, err)
		}
		if err := writer.Close(); err != nil {
			r.logger.Warningf("Failed to close writer for %s: %s", k, err)
		}
	}
	r.logger.Infof("Saved objects: %d, files: %d", len(objs), len(filesToLoad))
	return nil
}
