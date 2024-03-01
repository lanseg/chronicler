package storage

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"

	cm "github.com/lanseg/golang-commons/common"
	"github.com/lanseg/golang-commons/optional"
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
	return ol
}

func safeName(path string) string {
	safeName := nonAlphanumericRegex.ReplaceAllString(path, "_")
	if len(safeName) > 200 {
		safeName = safeName[len(safeName)-200:]
	}
	return safeName
}

func exists(path string) bool {
	_, err := os.Stat(path)
	return os.IsNotExist(err)
}

func write(path string, data []byte) optional.Optional[int] {
	return optional.OfError(len(data),
		os.WriteFile(path, data, os.ModePerm))
}

func writeObject(target string, object interface{}) optional.Optional[int] {
	return optional.MapErr(
		optional.OfError(json.Marshal(object)),
		func(json []byte) (int, error) {
			return len(json), os.WriteFile(target, json, os.ModePerm)
		})
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
	a, err := cm.FromJsonFile[Mapping](filepath.Join(o.root, mappingFile))
	if err != nil {
		o.logger.Warningf("Could not open mapping from %s: %s", path, err)
		return
	}
	o.mapping = a
	o.mapping.updateMappingIfNeeded()
	o.logger.Infof("Read %d record(s) from %s", len(a.Entities), path)
}

func (o *Overlay) ResolvePath(e *Entity) string {
	return filepath.Join(o.root, e.Name)
}

func (o *Overlay) Create(originalName string) optional.Optional[*Entity] {
	return o.Write(originalName, []byte{})
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

func (o *Overlay) Read(originalName string) optional.Optional[io.ReadCloser] {
	e := o.mapping.getEntity(originalName)
	if e == nil {
		return optional.OfError[io.ReadCloser](nil, fmt.Errorf("No entity with name %s", originalName))
	}
	f, err := os.Open(filepath.Join(o.root, e.Name))
	return optional.OfError[io.ReadCloser](io.ReadCloser(f), err)
}
