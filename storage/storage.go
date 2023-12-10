package storage

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"chronicler/downloader"
	"chronicler/records"
	rpb "chronicler/records/proto"
	"chronicler/util"
	"chronicler/webdriver"

	"github.com/lanseg/golang-commons/collections"
	cm "github.com/lanseg/golang-commons/common"
	"github.com/lanseg/golang-commons/optional"
)

const (
	recordsetFileName = "record.json"
	webdriverPort     = 2828
	firefoxProfile    = "/tmp/tmp.QTFqrzeJX4/"
)

type Storage interface {
	SaveRecords(r *rpb.RecordSet) error
	ListRecords() optional.Optional[[]*rpb.RecordSet]
	GetFile(id string, filename string) optional.Optional[[]byte]
}

type LocalStorage struct {
	Storage

	driver     *webdriver.ExclusiveWebDriver
	downloader downloader.Downloader
	overlay    *Overlay
	runner     *util.Runner

	logger  *cm.Logger
	modTime time.Time
	root    string

	recordCache map[string]*rpb.RecordSet
}

func fromJson[T any](bytes []byte) (*T, error) {
	result := new(T)
	err := json.Unmarshal(bytes, result)
	if err != nil {
		return nil, err
	}
	return result, err
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
		sDec, err := base64.StdEncoding.DecodeString(content)
		if err != nil {
			s.logger.Warningf("Could not decode base64: %s", err)
			return
		}
		written, err := s.getOverlay(id).Write(fname, sDec).Get()
		if err != nil {
			s.logger.Warningf("Could not write decoded base64: %s", err)
			return
		}
		s.logger.Debugf("Written %d byte(s) to file %v", written, fname)
	}
}

func (s *LocalStorage) getRecord(id string) optional.Optional[*rpb.RecordSet] {
	return optional.MapErr(
		s.getOverlay(id).Read(recordsetFileName),
		func(bytes []byte) (*rpb.RecordSet, error) {
			return fromJson[rpb.RecordSet](bytes)
		})
}

func (s *LocalStorage) savePageView(id string, url string) {
	s.driver.Batch(func(d webdriver.WebDriver) {
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

func (s *LocalStorage) SaveRecords(r *rpb.RecordSet) error {
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

		for _, link := range r.Links {
			if util.IsYoutubeLink(link) && strings.Contains(link, "v=") {
				s.logger.Debugf("Found youtube link: %s", link)
				if err := util.DownloadYoutube(link, s.root); err != nil {
					s.logger.Warningf("Failed to download youtube video: %s", err)
				}
			}

		}

		for _, file := range r.GetFiles() {
			if err := s.downloadFile(rs.Id, file.FileUrl); err != nil {
				continue
			}
		}

	}
	o := s.getOverlay(rs.Id)
	bytes, err := json.Marshal(rs)
	if err != nil {
		return err
	}

	_, err = o.Write(recordsetFileName, bytes).Get()
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

func (s *LocalStorage) ListRecords() optional.Optional[[]*rpb.RecordSet] {
	if s.isDirty() {
		s.refreshCache()
		s.modTime = time.Now()
	}
	values := collections.Values(s.recordCache)
	sort.Slice(values, func(i int, j int) bool {
		rsa := values[i]
		rsb := values[j]
		if len(rsa.Records) == 0 || len(rsb.Records) == 0 {
			return rsa.Id < rsb.Id
		}
		return values[i].Records[0].Time < values[j].Records[0].Time
	})
	return optional.Of(values)
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
	sort.Slice(result, func(i int, j int) bool {
		return result[i].Id < result[j].Id
	})
	return optional.Of(result)
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

func NewStorage(root string, driver *webdriver.ExclusiveWebDriver, downloader downloader.Downloader) Storage {
	log := cm.NewLogger("storage")
	log.Infof("Storage root set to \"%s\"", root)

	ls := &LocalStorage{
		root:       root,
		logger:     log,
		runner:     util.NewRunner(),
		driver:     driver,
		downloader: downloader,
	}
	ls.refreshCache()
	return ls
}
