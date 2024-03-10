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
	opt "github.com/lanseg/golang-commons/optional"
)

type localStorage struct {
	Storage

	overlay     *Overlay
	logger      *cm.Logger
	root        string
	recordCache map[string]*rpb.RecordSet
}

func (s *localStorage) getOverlay(id string) *Overlay {
	overlayRoot := filepath.Join(s.root, id)
	if s.overlay == nil || s.overlay.root != overlayRoot {
		s.overlay = NewOverlay(overlayRoot, cm.UUID4)
	}
	return s.overlay
}

func (s *localStorage) GetRecordSet(id string) opt.Optional[*rpb.RecordSet] {
	return opt.MapErr(
		opt.MapErr(s.GetFile(id, recordsetFileName), func(r io.ReadCloser) ([]byte, error) {
			defer r.Close()
			return io.ReadAll(r)
		}),
		cm.FromJson[rpb.RecordSet])
}

func (s *localStorage) GetFile(id string, filename string) opt.Optional[io.ReadCloser] {
	s.logger.Infof("GetFile %s %s", id, filename)
	return s.getOverlay(id).Read(filename)
}

func (s *localStorage) PutFile(id string, filename string, src io.Reader) error {
	s.logger.Infof("PutFile %s/%s", id, filename)
	return s.getOverlay(id).CopyFrom(filename, src)
}

func (s *localStorage) ListRecordSets() opt.Optional[[]*rpb.RecordSet] {
	return opt.Of(records.SortRecordSets(collections.Values(s.recordCache)))
}

func (s *localStorage) DeleteRecordSet(id string) error {
	if len(id) != 36 {
		return fmt.Errorf("Looks like uuid is incorrect: %q", id)
	}
	path := filepath.Join(s.root, id)
	if err := os.RemoveAll(filepath.Join(s.root, id)); err != nil {
		return err
	}
    delete(s.recordCache, id)
	s.logger.Debugf("Deleted recordset at %s", path)
	return nil
}

func (s *localStorage) SaveRecordSet(newSet *rpb.RecordSet) error {
	if newSet.Id == "" {
		return fmt.Errorf("Record without an id")
	}
	return s.writeRecordSet(
		records.MergeRecordSets(
			s.GetRecordSet(newSet.Id).OrElse(&rpb.RecordSet{}),
			newSet))
}

func (s *localStorage) getAllRecords() opt.Optional[[]*rpb.RecordSet] {
	result := []*rpb.RecordSet{}
	files, err := ioutil.ReadDir(s.root)
	if err != nil {
		return opt.OfError([]*rpb.RecordSet{}, err)
	}
	for _, f := range files {
		s.GetRecordSet(f.Name()).IfPresent(func(r *rpb.RecordSet) {
			result = append(result, r)
		})
	}
	return opt.Of(records.SortRecordSets(result))
}

func (s *localStorage) writeRecordSet(rs *rpb.RecordSet) error {
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

func (s *localStorage) refreshCache() {
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

func NewLocalStorage(root string) Storage {
	log := cm.NewLogger("storage")
	log.Infof("Storage root set to \"%s\"", root)

	ls := &localStorage{
		root:   root,
		logger: log,
	}
	ls.refreshCache()
	return ls
}
