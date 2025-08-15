package resolver

import (
	"fmt"
	"net/url"
	"path/filepath"
	"slices"
	"time"

	"chronicler/adapter"
	cm "chronicler/common"
	opb "chronicler/proto"
	"chronicler/storage"
)

const (
	objectFileName = "snapshot"
)

type Resolver struct {
	root     string
	adapters []adapter.Adapter
	loader   cm.Downloader
	logger   *cm.Logger
}

func NewResolver(root string, loader cm.Downloader, adapters []adapter.Adapter) *Resolver {
	return &Resolver{
		adapters: adapters,
		loader:   loader,
		root:     root,
		logger:   cm.NewLogger("Resolver"),
	}
}

func (r *Resolver) Resolve(link *opb.Link) error {
	for _, adapter := range r.adapters {
		if adapter.Match(link) {
			result, err := adapter.Get(link)
			if err != nil {
				return fmt.Errorf("adapter %q cannot get data from %q: %w", adapter, link, err)
			}
			return r.save(link, result)
		}
	}
	return fmt.Errorf("no adapter found for link %q", link)
}

func (r *Resolver) getStorage(link *opb.Link) (*storage.BlockStorage, error) {
	ls, err := storage.NewLocalStorage(filepath.Join(r.root, cm.UUID4For(link)))
	if err != nil {
		return nil, err
	}
	return &storage.BlockStorage{
		Storage: ls,
	}, nil
}

func (r *Resolver) save(link *opb.Link, objs []*opb.Object) error {
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
