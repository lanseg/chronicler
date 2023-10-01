package storage

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"

	"github.com/lanseg/golang-commons/optional"
    cm "github.com/lanseg/golang-commons/common" 
)

const (
	mappingFile = "mapping.json"
)

var nonAlphanumericRegex = regexp.MustCompile(`[^a-zA-Z0-9\\. ]+`)

type IdSource func() string

type Entity struct {
	Id           string `json:"id"`
	Name         string `json:"name"`
	OriginalName string `json:"original_name"`
}

type Mapping struct {
	Entities []*Entity `json:"entities"`

	entityByName map[string]*Entity
}

func (m *Mapping) updateMappingIfNeeded() {
	if m.Entities == nil {
		m.Entities = []*Entity{}
		m.entityByName = map[string]*Entity{}
	}
	if len(m.entityByName) == len(m.Entities) {
		return
	}
	m.entityByName = map[string]*Entity{}
	for _, e := range m.Entities {
		m.entityByName[e.OriginalName] = e
	}
}

func (m *Mapping) addEntity(e *Entity) {
	m.updateMappingIfNeeded()
	m.entityByName[e.OriginalName] = e
	m.Entities = append(m.Entities, e)
}

func (m *Mapping) getEntity(originalName string) *Entity {
	m.updateMappingIfNeeded()
	return m.entityByName[originalName]
}

type Overlay struct {
	logger *cm.Logger

	idSrc   IdSource
	mapping *Mapping
	root    string
}

func NewOverlay(root string, idSrc IdSource) *Overlay {
	ol := &Overlay{
		root:    root,
		idSrc:   idSrc,
		mapping: &Mapping{},
		logger:  cm.NewLogger("Overlay"),
	}
	if err := os.MkdirAll(root, os.ModePerm); err != nil {
		ol.logger.Warningf("Could not create directory at %s: %s", root, err)
	}
	ol.readMapping()
	ol.saveMapping()
	return ol
}

func safeName(path string) string {
	safeName := nonAlphanumericRegex.ReplaceAllString(path, "_")
	if len(safeName) > 200 {
		safeName = safeName[len(safeName)-200:]
	}
	return safeName
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

func (o *Overlay) Write(originalName string, bytes []byte) optional.Optional[*Entity] {
	name := safeName(originalName)
	return optional.Map(
		write(filepath.Join(o.root, name), bytes),
		func(int) *Entity {
			e := &Entity{o.idSrc(), name, originalName}
			o.mapping.addEntity(e)
			o.saveMapping()
			o.logger.Debugf("Saved %s to %s", originalName, name)
			return e
		})
}

func (o *Overlay) Read(originalName string) optional.Optional[[]byte] {
	e := o.mapping.getEntity(originalName)
	if e == nil {
		return optional.OfError([]byte{}, fmt.Errorf("No entity with name %s", originalName))
	}
	return read(filepath.Join(o.root, e.Name))
}
