package storage

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"sort"
	"strings"

	"chronicler/firefox"
	"chronicler/records"
	rpb "chronicler/records/proto"
	"chronicler/util"

	"github.com/lanseg/golang-commons/optional"
    cm "github.com/lanseg/golang-commons/common" 
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

	webdriver  firefox.WebDriver
	downloader *HttpDownloader
	overlay    *Overlay
	runner     *util.Runner

	recordCache map[string]string
	logger      *cm.Logger
	root        string
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

func (s *LocalStorage) savePageView(id string, url string) {
	s.webdriver.Navigate(url)
	s.webdriver.TakeScreenshot().IfPresent(
		s.saveBase64(id, "pageview_page.png"))
	s.webdriver.Print().IfPresent(
		s.saveBase64(id, "pageview_page.png"))
}

func (s *LocalStorage) downloadFile(id string, link string) error {
	o := s.getOverlay(id)
	data, err := s.downloader.Download(link)
	if err != nil {
		s.logger.Warningf("Failed to download file %s: %s", link, err)
		return err
	}
	_, err = o.Write(link, data).Get()
	return err
}

func (s *LocalStorage) SaveRecords(r *rpb.RecordSet) error {
	if r.Id == "" {
		return fmt.Errorf("Record without an id")
	}
	o := s.getOverlay(r.Id)
	mergedSet := records.MergeRecordSets(
		optional.MapErr(o.Read(recordsetFileName),
			fromJson[rpb.RecordSet],
		).OrElse(&rpb.RecordSet{}), r)
	return s.writeRecordSet(mergedSet)
}

func (s *LocalStorage) writeRecordSet(rs *rpb.RecordSet) error {
	if rs.Id == "" {
		return fmt.Errorf("Record must have an ID")
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
	for _, r := range rs.Records {
		if r.Source != nil && r.Source.Url != "" {
			s.savePageView(rs.Id, r.Source.Url)
			r.Files = append(r.Files, &rpb.File{
				FileId:   "page_view_png",
				LocalUrl: "pageview_page.png",
			}, &rpb.File{
				FileId:   "page_view_pdf",
				LocalUrl: "pageview_page.pdf",
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
	s.logger.Infof("Saved new record to %s", rs.Id)
	return nil
}

func (s *LocalStorage) GetFile(id string, filename string) optional.Optional[[]byte] {
	return s.getOverlay(id).Read(filename)
}

func (s *LocalStorage) ListRecords() optional.Optional[[]*rpb.RecordSet] {
	result := []*rpb.RecordSet{}
	files, err := ioutil.ReadDir(s.root)
	if err != nil {
		return optional.OfError([]*rpb.RecordSet{}, err)
	}
	for _, f := range files {
		optional.MapErr(
			s.getOverlay(f.Name()).Read(recordsetFileName),
			fromJson[rpb.RecordSet]).
			IfPresent(func(r *rpb.RecordSet) {
				result = append(result, r)
			})
	}
	sort.Slice(result, func(i int, j int) bool {
		return result[i].Id < result[j].Id
	})
	return optional.Of(result)
}

func NewStorage(root string, webdriver firefox.WebDriver) Storage {
	log := cm.NewLogger("storage")
	log.Infof("Storage root set to \"%s\"", root)

	if webdriver == nil {
		log.Infof("No webdriver provided, starting firefox")
		ff := firefox.StartFirefox(webdriverPort, firefoxProfile)
		webdriver = ff.Driver
	}
	webdriver.NewSession()

	return &LocalStorage{
		root:        root,
		logger:      log,
		runner:      util.NewRunner(),
		webdriver:   webdriver,
		downloader:  NewHttpDownloader(nil),
		recordCache: map[string]string{},
	}
}
