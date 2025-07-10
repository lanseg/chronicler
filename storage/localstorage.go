package storage

import (
	"chronicler/common"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"
)

const (
	maxBackups      = 1000
	maxNameLen      = 200
	defaultPerms    = 0777
	defaultMetadata = ".metadata"
	defaultSnapshot = ".snapshot"
	defaultMapping  = defaultMetadata + "/mapping.json"
)

type localStorage struct {
	Storage

	mux        sync.Mutex
	root       string
	localNames map[string]string
	logger     *common.Logger
}

func NewLocalStorage(root string) (Storage, error) {
	dir := filepath.Join(root, defaultMetadata)
	if err := os.MkdirAll(dir, defaultPerms); err != nil {
		return nil, errors.Join(fmt.Errorf("cannot create storage directory %q", dir), err)
	}
	storage := &localStorage{
		root:       root,
		localNames: map[string]string{},
		logger:     common.NewLogger("LocalStorage"),
	}
	if err := storage.readMapping(); err != nil {
		return nil, fmt.Errorf("cannot open storage mapping: %s", err)
	}
	return storage, nil
}

func (ls *localStorage) saveMapping() error {
	bytes, err := json.Marshal(ls.localNames)
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(ls.root, defaultMapping), bytes, defaultPerms)
}

func (ls *localStorage) readMapping() error {
	mapFile := filepath.Join(ls.root, defaultMapping)
	bytes, err := os.ReadFile(mapFile)
	if err != nil {
		if os.IsNotExist((err)) {
			return nil
		}
		return errors.Join(fmt.Errorf("cannot read storage mapping %q", mapFile), err)
	}
	mapping := map[string]string{}
	err = json.Unmarshal(bytes, &mapping)
	if err != nil {
		return err
	}

	ls.localNames = mapping
	return nil
}

func (ls *localStorage) snapshotFile(localName string) error {
	snapshotRoot := filepath.Join(ls.root, defaultSnapshot)
	if err := os.MkdirAll(snapshotRoot, defaultPerms); err != nil {
		return err
	}
	i := 0
	for ; i < maxBackups; i++ {
		backupName := filepath.Join(snapshotRoot, fmt.Sprintf("%s_%04d", localName, i))
		if _, err := os.Stat(backupName); errors.Is(err, os.ErrNotExist) {
			return os.Rename(filepath.Join(ls.root, localName), backupName)
		}
	}
	return fmt.Errorf("too many backups already")
}

func (ls *localStorage) Put(put *PutRequest) (io.WriteCloser, error) {
	ls.mux.Lock()
	defer ls.mux.Unlock()

	localName := common.Sanitize(put.Url, maxNameLen)
	localPath := filepath.Join(ls.root, localName)
	if _, err := os.Stat(localPath); err == nil {
		if put.SaveOnOverwrite {
			ls.logger.Debugf("File %q will be saved on overwrite", put.Url)
			if err = ls.snapshotFile(localName); err != nil {
				return nil, err
			}
		}
	}
	file, err := os.Create(localPath)
	if err != nil {
		return nil, fmt.Errorf("cannot open for writing %s/%s: %s", ls.root, put.Url, err)
	}
	ls.localNames[put.Url] = localName
	ls.saveMapping()
	return file, nil
}

func (ls *localStorage) Get(get *GetRequest) (io.ReadCloser, error) {
	ls.mux.Lock()
	defer ls.mux.Unlock()

	localName, ok := ls.localNames[get.Url]
	if !ok {
		return nil, fmt.Errorf("cannot open %s/%s: %s", ls.root, get.Url, os.ErrNotExist)
	}
	file, err := os.Open(filepath.Join(ls.root, localName))
	if err != nil {
		return nil, fmt.Errorf("cannot open %s/%s: %s", ls.root, get.Url, err)
	}
	return file, nil
}

func (ls *localStorage) List(list *ListRequest) (*ListResponse, error) {
	ls.mux.Lock()
	defer ls.mux.Unlock()

	result := &ListResponse{}
	snapshotRoot := filepath.Join(ls.root, defaultSnapshot)
	for actual, local := range ls.localNames {
		if len(list.Url) > 0 {
			found := false
			for _, url := range list.Url {
				if url == actual {
					found = true
					break
				}
			}
			if !found {
				continue
			}
		}
		item := StorageItem{
			Url: actual,
		}
		if list.WithSnapshots {
			for i := 0; i < maxBackups; i++ {
				backupName := filepath.Join(snapshotRoot, fmt.Sprintf("%s_%04d", local, i))
				if _, err := os.Stat(backupName); errors.Is(err, os.ErrNotExist) {
					break
				}
				item.Versions = append(item.Versions, fmt.Sprintf("%04d", i))
			}
		}
		result.Items = append(result.Items, item)
	}
	return result, nil
}
