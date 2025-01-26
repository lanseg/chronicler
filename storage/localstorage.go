package storage

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

const (
	defaultPerms    = 0777
	defaultMetadata = ".metadata"
	defaultMapping  = defaultMetadata + "/mapping.json"
)

func sanitizeUrl(remotePath string) string {
	builder := strings.Builder{}
	for _, r := range remotePath {
		if (r >= 'a' && r <= 'z') ||
			(r >= 'A' && r <= 'Z') ||
			(r >= '0' && r <= '9') ||
			r == '-' || r == '_' || r == '.' {
			builder.WriteRune(r)
		} else {
			builder.WriteRune('_')
		}
	}
	return builder.String()
}

type localStorage struct {
	Storage

	writeMux   sync.Mutex
	root       string
	localNames map[string]string
}

func NewLocalStorage(root string) (Storage, error) {
	if err := os.MkdirAll(root, defaultPerms); err != nil {
		return nil, err
	}
	if err := os.MkdirAll(filepath.Join(root, defaultMetadata), defaultPerms); err != nil {
		return nil, err
	}
	storage := &localStorage{
		root:       root,
		localNames: map[string]string{},
	}
	if err := storage.readMapping(); err != nil {
		return nil, err
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
	bytes, err := os.ReadFile(filepath.Join(ls.root, defaultMapping))
	if err != nil {
		if os.IsNotExist((err)) {
			return nil
		}
		return err
	}
	mapping := map[string]string{}
	err = json.Unmarshal(bytes, &mapping)
	if err != nil {
		return err
	}

	ls.localNames = mapping
	return nil
}

func (ls *localStorage) Put(put *PutRequest) (io.WriteCloser, error) {
	ls.writeMux.Lock()
	defer ls.writeMux.Unlock()

	localName := sanitizeUrl(put.Url)
	file, err := os.Create(filepath.Join(ls.root, localName))
	if err != nil {
		return nil, fmt.Errorf("Cannot open for writing %s/%s: %s", ls.root, put.Url, err)
	}
	ls.localNames[put.Url] = localName
	ls.saveMapping()
	return file, nil
}

func (ls *localStorage) Get(get *GetRequest) (io.ReadCloser, error) {
	localName, ok := ls.localNames[get.Url]
	if !ok {
		return nil, fmt.Errorf("Cannot open %s/%s: %s", ls.root, get.Url, os.ErrNotExist)
	}
	file, err := os.Open(filepath.Join(ls.root, localName))
	if err != nil {
		return nil, fmt.Errorf("Cannot open %s/%s: %s", ls.root, get.Url, err)
	}
	return file, nil
}

func (ls *localStorage) List(list *ListRequest) (*ListResponse, error) {
	result := &ListResponse{}
	for k := range ls.localNames {
		result.Url = append(result.Url, k)
	}
	return result, nil
}
