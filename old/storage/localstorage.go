package storage

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"

	aio "github.com/lanseg/golang-commons/almostio"
	col "github.com/lanseg/golang-commons/collections"
	cm "github.com/lanseg/golang-commons/common"
	opt "github.com/lanseg/golang-commons/optional"

	"chronicler/records"
	rpb "chronicler/records/proto"
)

type localStorage struct {
	Storage

	ovlmux      sync.Mutex
	overlays    sync.Map
	recordCache sync.Map

	logger *cm.Logger
	root   string
}

func (s *localStorage) getOverlay(id string) aio.Overlay {
	result, _ := s.overlays.LoadOrStore(id, func() aio.Overlay {
		ovl, err := aio.NewLocalOverlay(
			filepath.Join(s.root, id),
			aio.NewJsonMarshal[aio.OverlayMetadata]())
		if err != nil {
			s.logger.Warningf("Cannot open overlay for %s: %s", id, err)
		}
		return ovl
	}())
	return result.(aio.Overlay)
}

func (s *localStorage) attachMetadata(rs *rpb.RecordSet) *rpb.RecordSet {
	fileNames := map[string]*aio.FileMetadata{}
	for _, r := range rs.GetRecords() {
		for _, f := range r.GetFiles() {
			fileNames[f.GetFileUrl()] = nil
		}
	}
	ovl := s.getOverlay(rs.GetId())
	for _, md := range ovl.GetMetadata(col.Keys(fileNames)) {
		if md == nil {
			continue
		}
		fileNames[md.Name] = md
	}
	rs.FileMetadata = []*rpb.FileMetadata{}
	for name, meta := range fileNames {
		if meta == nil {
			continue
		}
		rs.FileMetadata = append(rs.FileMetadata, &rpb.FileMetadata{
			Name:     name,
			Mimetype: meta.Mime,
			Checksum: fmt.Sprintf("sha256/%s", meta.Sha256),
		})
	}
	return rs
}

func (s *localStorage) GetRecordSet(id string) opt.Optional[*rpb.RecordSet] {
	return opt.Map(
		opt.MapErr(
			opt.MapErr(s.GetFile(id, recordsetFileName), func(r io.ReadCloser) ([]byte, error) {
				defer r.Close()
				return io.ReadAll(r)
			}),
			cm.FromJson[rpb.RecordSet]), s.attachMetadata)
}

func (s *localStorage) GetFile(id string, filename string) opt.Optional[io.ReadCloser] {
	s.logger.Infof("GetFile %s %s", id, filename)
	return opt.OfError(s.getOverlay(id).OpenRead(filename))
}

func (s *localStorage) PutFile(id string, filename string, src io.Reader) error {
	s.logger.Infof("PutFile %s/%s", id, filename)
	wc, err := s.getOverlay(id).OpenWrite(filename)
	if err != nil {
		return err
	}
	defer wc.Close()
	_, err = io.Copy(wc, src)
	return err
}

func (s *localStorage) ListRecordSets(request *rpb.ListRecordsRequest) opt.Optional[[]*rpb.RecordSet] {
	sorting := &rpb.Sorting{Field: rpb.Sorting_CREATE_TIME, Order: rpb.Sorting_ASC}
	paging := &rpb.Paging{Offset: 0, Size: 20}
	if request != nil {
		if request.Sorting != nil {
			sorting = request.Sorting
		}
		if request.Paging != nil {
			paging = request.Paging
		}
	}
	values := []*rpb.RecordSet{}
	s.recordCache.Range(func(key, value any) bool {
		if request.Query == "" || strings.Contains(fmt.Sprintf("%s", value), request.Query) {
			values = append(values, value.(*rpb.RecordSet))
		}
		return true
	})
	sorted := records.SortRecordSets(values, sorting)
	return opt.Of(sorted[min(len(sorted), int(paging.Offset)):min(len(sorted), int(paging.Offset+paging.Size))])
}

func (s *localStorage) DeleteRecordSet(id string) error {
	if len(id) != 36 {
		return fmt.Errorf("Looks like uuid is incorrect: %q", id)
	}
	s.ovlmux.Lock()
	defer s.ovlmux.Unlock()
	path := filepath.Join(s.root, id)
	if err := os.RemoveAll(filepath.Join(s.root, id)); err != nil {
		return err
	}
	s.recordCache.Delete(id)
	s.logger.Debugf("Deleted recordset at %s", path)
	return nil
}

func (s *localStorage) SaveRecordSet(newSet *rpb.RecordSet) error {
	if newSet.Id == "" {
		return fmt.Errorf("Record without an id")
	}
	s.ovlmux.Lock()
	defer s.ovlmux.Unlock()
	return s.writeRecordSet(
		records.MergeRecordSets(
			s.GetRecordSet(newSet.Id).OrElse(&rpb.RecordSet{}),
			newSet))
}

func (s *localStorage) getAllRecords() opt.Optional[[]*rpb.RecordSet] {
	result := []*rpb.RecordSet{}
	files, err := os.ReadDir(s.root)
	if err != nil {
		return opt.OfError([]*rpb.RecordSet{}, err)
	}
	for _, f := range files {
		s.GetRecordSet(f.Name()).IfPresent(func(r *rpb.RecordSet) {
			result = append(result, r)
		})
	}
	return opt.Of(result)
}

func (s *localStorage) writeRecordSet(rs *rpb.RecordSet) error {
	bytes, err := json.Marshal(rs)
	if err != nil {
		return err
	}

	wc, err := s.getOverlay(rs.Id).OpenWrite(recordsetFileName)
	if err != nil {
		return err
	}
	wc.Write(bytes)

	s.recordCache.Store(rs.Id, rs)
	s.logger.Infof("Saved new record to %s", rs.Id)
	return wc.Close()
}

func (s *localStorage) refreshCache() {
	s.logger.Infof("Refreshing cache")
	s.getAllRecords().IfPresent(func(allRecords []*rpb.RecordSet) {
		s.logger.Infof("Found %d records", len(allRecords))
		for _, r := range allRecords {
			s.recordCache.Store(r.Id, r)
		}
	})
}

func NewLocalStorage(root string) Storage {
	log := cm.NewLogger("storage")
	log.Infof("Storage root set to \"%s\"", root)

	ls := &localStorage{
		root:        root,
		logger:      log,
		recordCache: sync.Map{},
		overlays:    sync.Map{},
	}
	ls.refreshCache()
	return ls
}
