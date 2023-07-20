package storage

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"chronicler/util"

	"github.com/lanseg/golang-commons/optional"
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

func (f *RelativeFS) Read(path string) optional.Optional[[]byte] {
	return optional.OfError(os.ReadFile(f.Resolve(path)))
}

func (f *RelativeFS) ListFiles(path string) ([]os.FileInfo, error) {
	return ioutil.ReadDir(f.Resolve("."))
}

func ReadJSON[T any](f *RelativeFS, path string) optional.Optional[*T] {
	return optional.MapErr(f.Read(path), func(bytes []byte) (*T, error) {
		result := new(T)
		err := json.Unmarshal(bytes, result)
		return result, err
	})
}

func NewRelativeFS(root string) *RelativeFS {
	return &RelativeFS{
		logger: util.NewLogger(fmt.Sprintf("RelativeFS %s", root)),
		root:   root,
	}
}
