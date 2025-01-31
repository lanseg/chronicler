package resolver

import (
	"chronicler/adapter"
	"chronicler/common"
	opb "chronicler/proto"
	"chronicler/storage"
	"net/url"
	"path/filepath"
	"sync"
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

	taskWaiter sync.WaitGroup

	done     chan bool
	tasks    chan resolverTask
	loader   common.Downloader
	root     string
	adapters []adapter.Adapter
	logger   *common.Logger
}

func NewResolver(root string, loader common.Downloader, adapters []adapter.Adapter) Resolver {
	r := &resolver{
		taskWaiter: sync.WaitGroup{},

		done:     make(chan bool, 1),
		tasks:    make(chan resolverTask, 10),
		adapters: adapters,
		loader:   loader,
		root:     root,
		logger:   common.NewLogger("Resolver"),
	}
	r.logger.Infof("Initialized resolver with %d adapters", len(adapters))
	return r
}

func (r *resolver) Start() {
	r.logger.Infof("Starting resolver thread")
	go func() {
	loop:
		for {
			select {
			case <-r.done:
				break loop
			case task := <-r.tasks:
				r.resolveTask(task)
				r.taskWaiter.Done()
			}
		}
		close(r.tasks)
		close(r.done)
	}()
}

func (r *resolver) Wait() {
	r.logger.Infof("Waiting for all tasks to complete")
	r.taskWaiter.Wait()
}

func (r *resolver) Stop() {
	r.logger.Infof("Stopping resolver")
	r.done <- true
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
	ls, err := storage.NewLocalStorage(filepath.Join(r.root, common.UUID4For(link)))
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
		return err
	}
	s, err := r.getStorage(link)
	if err != nil {
		return err
	}
	bytesWritten, err := s.PutObject(&storage.PutRequest{Url: "objects.json"}, objs)
	if err != nil {
		return err
	}
	r.logger.Infof("Saved objects.json, written bytes: %d", bytesWritten)

	filesToLoad := map[*url.URL]bool{}
	for _, obj := range objs {
		for _, attachment := range obj.Attachment {
			if attachment.Mime == "" {
				continue
			}
			fileUrl, err := url.Parse(attachment.Url)
			if err != nil {
				r.logger.Warningf("Cannot parse url \"%s\" from object %s: %s", obj.Id, fileUrl, err)
				continue
			}
			filesToLoad[fileUrl] = true
		}
	}

	file := 0
	toLoad := len(filesToLoad)
	r.logger.Infof("Files to download: %d", toLoad)
	for k := range filesToLoad {
		file += 1
		r.logger.Infof("Downloading [%d of %d] %s", file, toLoad, k)
		size, err := r.loader.Download(k.String(), s)
		if err != nil {
			r.logger.Warningf("Failed to download %s: %s", k, err)
			continue
		}
		if size == -1 {
			r.logger.Infof("No need to download file %s", k)
		} else {
			r.logger.Infof("Downloaded %s, size %d", k, size)
		}
	}
	r.logger.Infof("Saved objects: %d, files: %d", len(objs), len(filesToLoad))
	return nil
}
