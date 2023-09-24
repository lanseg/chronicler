package storage

import (
	"encoding/json"
	"os"
	"path/filepath"

	"chronicler/util"

	"github.com/lanseg/golang-commons/optional"
)

const (
	mappingFile = "mapping.json"
)

type IdSource func() string

type Entity struct {
	Id         string `json:"id"`
	LocalName  string `json:"local_name"`
	ActualName string `json:"actual_name"`
}

type Mapping struct {
	Entities []*Entity `json:"entities"`

	entityByName map[string]*Entity
}

func (m *Mapping) updateMappingIfNeeded() {
	if len(m.entityByName) == len(m.Entities) {
		return
	}
	m.entityByName = map[string]*Entity{}
	for _, e := range m.Entities {
		m.entityByName[e.LocalName] = e
	}
}

func (m *Mapping) getEntity(realName string) *Entity {
	m.updateMappingIfNeeded()
	return m.entityByName[realName]
}

type Overlay struct {
	logger *util.Logger

	idSrc   IdSource
	mapping *Mapping
	root    string
}

func NewOverlay(root string, idSrc IdSource) *Overlay {
	ol := &Overlay{
		root:   root,
		idSrc:  idSrc,
		logger: util.NewLogger("Overlay"),
	}
	if err := os.MkdirAll(root, os.ModePerm); err != nil {
		ol.logger.Warningf("Could not create directory at %s: %s", root, err)
	}
	ol.readMapping()
	ol.saveMapping()
	return ol
}

func read(path string) optional.Optional[[]byte] {
	return optional.OfError(os.ReadFile(path))
}

func write(path string, data []byte) optional.Optional[int] {
	return optional.OfError(len(data),
		os.WriteFile(path, data, os.ModePerm))
}

func readObject[T any](target string) optional.Optional[*T] {
	return optional.MapErr(read(target), func(data []byte) (*T, error) {
		result := new(T)
		err := json.Unmarshal(data, result)
		return result, err
	})
}

func writeObject(target string, object interface{}) optional.Optional[int] {
	bytes, err := json.Marshal(object)
	if err != nil {
		return optional.OfError(0, err)
	}
	return write(target, bytes)
}

func (o *Overlay) getMappingPath() string {
	return filepath.Join(o.root, mappingFile)
}

func (o *Overlay) saveMapping() {
	path := o.getMappingPath()
	o.logger.Infof("Saving mapping to %s", path)
	writeObject(path, o.mapping).IfPresent(func(bytes int) {
		o.logger.Infof("Saved %d bytes to %s", bytes, path)
	})
}

func (o *Overlay) readMapping() {
	path := o.getMappingPath()
	o.logger.Infof("Reading mapping from %s", path)
	a, err := readObject[Mapping](filepath.Join(o.root, mappingFile)).Get()
	if err != nil {
		o.logger.Warningf("Could not open mapping from %s: %s", path, err)
		return
	}
	o.mapping = a
	o.logger.Infof("Read %d record(s) from %s", len(a.Entities), path)
}
