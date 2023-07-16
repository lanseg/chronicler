package storage

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"chronicler/util"
)

type RelativeFS struct {
	logger *util.Logger

	root string
}

func (f *RelativeFS) Resolve(relativePath string) string {
	return filepath.Join(f.root, relativePath)
}

func (f *RelativeFS) MkDir(path string) error {
	recordRoot := f.Resolve(path)
	f.logger.Debugf("Creating directory at [%s]/%s: %s", f.root, path, recordRoot)
	return os.MkdirAll(recordRoot, os.ModePerm)
}

func (f *RelativeFS) Write(path string, value []byte) error {
	return os.WriteFile(f.Resolve(path), value, os.ModePerm)
}

func (f *RelativeFS) WriteJSON(path string, data interface{}) error {
	bytes, err := json.Marshal(data)
	if err != nil {
		return err
	}
	return f.Write(path, bytes)
}

func (f *RelativeFS) Read(path string) ([]byte, error) {
	b, err := os.ReadFile(f.Resolve(path))
	if err != nil {
		f.logger.Warningf("Error reading file: %s", err)
		return nil, err
	}
	return b, err
}

func ReadJson[T any](f *RelativeFS, path string) (*T, error) {
	result := new(T)
	bytes, err := f.Read(path)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(bytes, result)
	return result, err
}

func NewRelativeFS(root string) *RelativeFS {
	return &RelativeFS{
		logger: util.NewLogger(fmt.Sprintf("RelativeFS %s", root)),
		root:   root,
	}
}
