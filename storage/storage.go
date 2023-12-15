package storage

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"

	"chronicler/downloader"
	"chronicler/records"
	rpb "chronicler/records/proto"
	"chronicler/webdriver"

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
	DeleteRecordSet(id string) error
	GetFile(id string, filename string) optional.Optional[[]byte]
}

type LocalStorage struct {
	Storage

	browser    webdriver.Browser
	downloader downloader.Downloader
	overlay    *Overlay

	logger  *cm.Logger
	modTime time.Time
	root    string

	recordCache map[string]*rpb.RecordSet
}

func (s *LocalStorage) getOverlay(id string) *Overlay {
	overlayRoot := filepath.Join(s.root, id)
	if s.overlay == nil || s.overlay.root != overlayRoot {
		s.overlay = NewOverlay(overlayRoot, cm.UUID4)
	}
	return s.overlay
}

func (s *LocalStorage) saveBase64(id string, fname string) func(string) {
	return func(content string) {
		e, err := optional.MapErr(
			optional.OfError(base64.StdEncoding.DecodeString(content)),
			func(sDec []byte) (*Entity, error) {
				return s.getOverlay(id).Write(fname, sDec).Get()
			}).Get()
		if err == nil {
			s.logger.Debugf("Written %v to file %v", e.OriginalName, e.Name)
		} else {
			s.logger.Warningf("Could not save base64 data to %q: %v", fname, err)
		}
	}
}

func (s *LocalStorage) getRecord(id string) optional.Optional[*rpb.RecordSet] {
	return optional.MapErr(
		s.getOverlay(id).Read(recordsetFileName), cm.FromJson[rpb.RecordSet])
}

func (s *LocalStorage) savePageView(id string, url string) {
	s.browser.RunSession(func(d webdriver.WebDriver) {
		d.Navigate(url)
		d.TakeScreenshot().IfPresent(s.saveBase64(id, "pageview_page.png"))
		d.Print().IfPresent(s.saveBase64(id, "pageview_page.pdf"))
		d.GetPageSource().IfPresent(func(src string) {
			s.getOverlay(id).Write("pageview_page.html", []byte(src))
		})
	})
}

func (s *LocalStorage) downloadFile(id string, link string) error {
	o := s.getOverlay(id)
	path, err := optional.Map(o.Create(link), o.ResolvePath).Get()
	if err != nil {
		return err
	}
	if err = s.downloader.ScheduleDownload(link, path); err != nil {
		s.logger.Warningf("Failed to download file %s: %s", link, err)
		return err
	}
	return nil
}

func (s *LocalStorage) SaveRecordSet(r *rpb.RecordSet) error {
	if r.Id == "" {
		return fmt.Errorf("Record without an id")
	}
	mergedSet := records.MergeRecordSets(s.getRecord(r.Id).OrElse(&rpb.RecordSet{}), r)
	s.touch()
	return s.writeRecordSet(mergedSet)
}

func (s *LocalStorage) writeRecordSet(rs *rpb.RecordSet) error {
	if rs.Id == "" {
		return fmt.Errorf("Record must have an ID")
	}
	for _, r := range rs.Records {
		if r.Source != nil && r.Source.Url != "" {
			s.savePageView(rs.Id, r.Source.Url)
			r.Files = append(r.Files, &rpb.File{
				FileId:   "page_view_png",
				LocalUrl: "pageview_page.png",
			}, &rpb.File{
				FileId:   "page_view_pdf",
				LocalUrl: "pageview_page.pdf",
			}, &rpb.File{
				FileId:   "page_view_html",
				LocalUrl: "pageview_page.html",
			})
		}

		for _, file := range r.GetFiles() {
			if err := s.downloadFile(rs.Id, file.FileUrl); err != nil {
				s.logger.Warningf("Could not download file %q: %s", file.FileUrl, err)
			}
		}
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

func (s *LocalStorage) GetFile(id string, filename string) optional.Optional[[]byte] {
	s.logger.Infof("GetFile %s %s", id, filename)
	return s.getOverlay(id).Read(filename)
}

func (s *LocalStorage) ListRecordSets() optional.Optional[[]*rpb.RecordSet] {
	if s.isDirty() {
		s.refreshCache()
		s.modTime = time.Now()
	}
	return optional.Of(records.SortRecordSets(collections.Values(s.recordCache)))
}

func (s *LocalStorage) getAllRecords() optional.Optional[[]*rpb.RecordSet] {
	result := []*rpb.RecordSet{}
	files, err := ioutil.ReadDir(s.root)
	if err != nil {
		return optional.OfError([]*rpb.RecordSet{}, err)
	}
	for _, f := range files {
		s.getRecord(f.Name()).IfPresent(func(r *rpb.RecordSet) {
			result = append(result, r)
		})
	}
	return optional.Of(records.SortRecordSets(result))
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

func (s *LocalStorage) touch() {
	os.WriteFile(filepath.Join(".storage_dirty"), []byte{}, os.ModePerm)
	s.modTime = time.Now()
}

func (s *LocalStorage) isDirty() bool {
	stat, err := os.Stat(filepath.Join(".storage_dirty"))
	return os.IsNotExist(err) || stat.ModTime().After(s.modTime)
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

func NewStorage(root string, browser webdriver.Browser, downloader downloader.Downloader) Storage {
	log := cm.NewLogger("storage")
	log.Infof("Storage root set to \"%s\"", root)

	ls := &LocalStorage{
		root:       root,
		logger:     log,
		browser:    browser,
		downloader: downloader,
	}
	ls.refreshCache()
	return ls
}
