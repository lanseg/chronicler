package storage

import (
	"chronicler/common"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"
)

const (
	defaultMaxBackups = 5
	defaultMaxNameLen = 200
	defaultPerms      = 0744
	defaultMetadata   = ".metadata"
	defaultSnapshot   = ".snapshot"
	defaultMapping    = defaultMetadata + "/mapping.json"
)

type TimeSource func() time.Time

type localStorage struct {
	Storage

	maxBackups int
	maxNameLen int

	timeSource func() time.Time
	mux        sync.Mutex
	root       string
	localNames map[string]string
	logger     *common.Logger
}

type LocalStorageBuilder struct {
	StorageBuilder

	MaxBackups int
	MaxNameLen int
	Root       string
}

func (ls *LocalStorageBuilder) Build() (Storage, error) {
	if ls.Root == "" {
		return nil, fmt.Errorf("no root configured for the local builder")
	}
	result := &localStorage{
		root:       ls.Root,
		maxBackups: defaultMaxBackups,
		maxNameLen: defaultMaxNameLen,
		timeSource: time.Now,
		localNames: map[string]string{},
		logger:     common.NewLogger("LocalStorage"),
	}
	if ls.MaxBackups != 0 {
		result.maxBackups = ls.MaxBackups
	}
	if ls.MaxNameLen != 0 {
		result.maxNameLen = ls.MaxNameLen
	}
	return result, nil
}

func NewLocalStorage(root string) (Storage, error) {
	result, err := (&LocalStorageBuilder{Root: root}).Build()
	if err != nil {
		return nil, fmt.Errorf("cannot create local storage at root %q: %w", root, err)
	}
	dir := filepath.Join(root, defaultMetadata)
	if err := os.MkdirAll(dir, defaultPerms); err != nil {
		return nil, fmt.Errorf("cannot create storage directory %q: %w", dir, err)
	}

	if err := result.(*localStorage).readMapping(); err != nil {
		return nil, fmt.Errorf("cannot open storage mapping: %s", err)
	}
	return result, nil
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

func (ls *localStorage) getSnapshots(localName string) ([]string, error) {
	snapshotRoot := filepath.Join(ls.root, defaultSnapshot)
	files, err := os.ReadDir(snapshotRoot)
	if err != nil {
		return nil, err
	}
	versions := []string{}
	for _, f := range files {
		if strings.HasSuffix(f.Name(), localName) {
			versions = append(versions, f.Name())
		}
	}

	sort.Strings(versions)
	return versions, nil
}

func (ls *localStorage) snapshotFile(localName string) error {
	snapshotRoot := filepath.Join(ls.root, defaultSnapshot)
	if err := os.MkdirAll(snapshotRoot, defaultPerms); err != nil {
		return fmt.Errorf("cannot create folder for snapshots %q: %w", snapshotRoot, err)
	}

	versions, err := ls.getSnapshots(localName)
	if err != nil {
		return fmt.Errorf("cannot list snapshots in %q: %w", snapshotRoot, err)
	}
	for len(versions) > ls.maxBackups {
		if err := os.Remove(versions[0]); err != nil {
			return fmt.Errorf("cannot remove old snapshot %q: %w", versions[0], err)
		}
		versions = versions[1:]
	}

	timeStr := ls.timeSource().Format("2006-01-02_15-04-05.000")
	for v := range ls.maxBackups {
		backupName := filepath.Join(snapshotRoot, fmt.Sprintf("%s_%d_%s", timeStr, v, localName))
		if _, err := os.Stat(backupName); err != nil && os.IsNotExist(err) {
			return os.Rename(filepath.Join(ls.root, localName), backupName)
		}
	}
	return fmt.Errorf("still cannot create snapshot")
}

func (ls *localStorage) Put(put *PutRequest) (io.WriteCloser, error) {
	ls.mux.Lock()
	defer ls.mux.Unlock()

	localName := common.Sanitize(put.Url, ls.maxNameLen)
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
		item := StorageItem{Url: actual}
		if list.WithSnapshots {
			versions, err := ls.getSnapshots(local)
			fmt.Println("HERE: ", versions)
			if err != nil {
				ls.logger.Warningf("cannot get snapshots for %q: %s", local, err)
				continue
			}
			item.Versions = append(item.Versions, versions...)
		}
		result.Items = append(result.Items, item)
	}
	return result, nil
}
