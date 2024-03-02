package storage

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"

	"chronicler/records"
	rpb "chronicler/records/proto"

	"github.com/lanseg/golang-commons/collections"
	cm "github.com/lanseg/golang-commons/common"
	"github.com/lanseg/golang-commons/optional"
)

const (
	recordsetFileName = "record.json"
)

type Storage interface {
	SaveRecordSet(r *rpb.RecordSet) error
	ListRecordSets() optional.Optional[[]*rpb.RecordSet]
	GetRecordSet(id string) optional.Optional[*rpb.RecordSet]
	DeleteRecordSet(id string) error

	GetFile(id string, filename string) optional.Optional[io.ReadCloser]
	PutFile(id string, filename string, src io.Reader) error
}

type LocalStorage struct {
	Storage

	overlay     *Overlay
	logger      *cm.Logger
	root        string
	recordCache map[string]*rpb.RecordSet
}

func (s *LocalStorage) getOverlay(id string) *Overlay {
	overlayRoot := filepath.Join(s.root, id)
	if s.overlay == nil || s.overlay.root != overlayRoot {
		s.overlay = NewOverlay(overlayRoot, cm.UUID4)
	}
	return s.overlay
}

func (s *LocalStorage) GetRecordSet(id string) optional.Optional[*rpb.RecordSet] {
	return optional.MapErr(
		optional.MapErr(s.GetFile(id, recordsetFileName), func(r io.ReadCloser) ([]byte, error) {
			defer r.Close()
			return io.ReadAll(r)
		}),
		cm.FromJson[rpb.RecordSet])
}

func (s *LocalStorage) GetFile(id string, filename string) optional.Optional[io.ReadCloser] {
	s.logger.Infof("GetFile %s %s", id, filename)
	return s.getOverlay(id).Read(filename)
}

func (s *LocalStorage) PutFile(id string, filename string, src io.Reader) error {
	s.logger.Infof("PutFile %s/%s", id, filename)
	return s.getOverlay(id).CopyFrom(filename, src)
}

func (s *LocalStorage) ListRecordSets() optional.Optional[[]*rpb.RecordSet] {
	return optional.Of(records.SortRecordSets(collections.Values(s.recordCache)))
}

func (s *LocalStorage) DeleteRecordSet(id string) error {
	if len(id) != 36 {
		return fmt.Errorf("Looks like uuid is incorrect: %q", id)
	}
	path := filepath.Join(s.root, id)
	if err := os.RemoveAll(filepath.Join(s.root, id)); err != nil {
		return err
	}
	s.logger.Debugf("Deleted recordset at %s", path)
	s.refreshCache()
	return nil
}

func (s *LocalStorage) SaveRecordSet(r *rpb.RecordSet) error {
	if r.Id == "" {
		return fmt.Errorf("Record without an id")
	}
	return s.writeRecordSet(records.MergeRecordSets(s.GetRecordSet(r.Id).OrElse(&rpb.RecordSet{}), r))
}

func (s *LocalStorage) getAllRecords() optional.Optional[[]*rpb.RecordSet] {
	result := []*rpb.RecordSet{}
	files, err := ioutil.ReadDir(s.root)
	if err != nil {
		return optional.OfError([]*rpb.RecordSet{}, err)
	}
	for _, f := range files {
		s.GetRecordSet(f.Name()).IfPresent(func(r *rpb.RecordSet) {
			result = append(result, r)
		})
	}
	return optional.Of(records.SortRecordSets(result))
}

func (s *LocalStorage) writeRecordSet(rs *rpb.RecordSet) error {
	if rs.Id == "" {
		return fmt.Errorf("Record must have an ID")
	}

	bytes, err := json.Marshal(rs)
	if err != nil {
		return err
	}

	_, err = s.getOverlay(rs.Id).Write(recordsetFileName, bytes).Get()
	if err != nil {
		return err
	}

	s.recordCache[rs.Id] = rs
	s.logger.Infof("Saved new record to %s", rs.Id)
	return nil
}

func (s *LocalStorage) refreshCache() {
	s.logger.Infof("Refreshing cache")
	s.recordCache = map[string]*rpb.RecordSet{}
	s.getAllRecords().IfPresent(func(allRecords []*rpb.RecordSet) {
		s.logger.Infof("Found %d records", len(allRecords))
		for k, v := range collections.GroupBy(allRecords, func(rs *rpb.RecordSet) string {
			return rs.Id
		}) {
			s.recordCache[k] = v[0]
		}
	})

}

func NewStorage(root string) Storage {
	log := cm.NewLogger("storage")
	log.Infof("Storage root set to \"%s\"", root)

	ls := &LocalStorage{
		root:   root,
		logger: log,
	}
	ls.refreshCache()
	return ls
}
