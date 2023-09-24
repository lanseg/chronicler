package storage

import (
	"encoding/base64"
	"fmt"
	"net/url"
	"path"
	"path/filepath"
	"sort"
	"strings"

	"chronicler/firefox"
	"chronicler/records"
	rpb "chronicler/records/proto"
	"chronicler/util"

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

	webdriver  firefox.WebDriver
	downloader *HttpDownloader
	fs         *RelativeFS
	runner     *util.Runner

	recordCache map[string]string
	logger      *util.Logger
	root        string
}

func (s *LocalStorage) saveBase64(fname string) func(string) {
	return func(content string) {
		sDec, err := base64.StdEncoding.DecodeString(content)
		if err != nil {
			s.logger.Warningf("Could not decode base64: %s", err)
			return
		}
		err = s.fs.Write(fname, sDec)
		if err != nil {
			s.logger.Warningf("Could not write decoded base64: %s", err)
			return
		}
		s.logger.Debugf("Written %d byte(s) to file %s", len(sDec), fname)
	}
}

func (s *LocalStorage) savePageView(url string, target string) {
	s.webdriver.Navigate(url)
	s.webdriver.TakeScreenshot().IfPresent(s.saveBase64(
		filepath.Join(target, "page.png")))
	s.webdriver.Print().IfPresent(s.saveBase64(
		filepath.Join(target, "page.pdf")))
}

func (s *LocalStorage) downloadFile(root string, file *rpb.File) error {
	fileUrl, err := url.Parse(file.GetFileUrl())
	if err != nil || fileUrl.String() == "" {
		s.logger.Warningf("Malformed url for file: %s", file)
		return err
	}
	fname := path.Base(fileUrl.Path)
	local := s.fs.Resolve(filepath.Join(root, fname))
	if err := s.downloader.Download(file.GetFileUrl(), local); err != nil {
		s.logger.Warningf("Failed to download file %s to %s: %s", file, local, err)
		return err
	}
	file.LocalUrl = local
	return nil
}

func (s *LocalStorage) refreshCache() error {
	s.logger.Debugf("Refreshing record cache")
	s.recordCache = map[string]string{}
	files, err := s.fs.ListFiles("")
	if err != nil {
		return err
	}
	for _, info := range files {
		r, err := ReadJSON[rpb.RecordSet](s.fs, filepath.Join(info.Name(), recordsetFileName)).Get()
		if err != nil {
			s.logger.Warningf("Broken record, cannot parse %s: %s", info.Name(), err)
			continue
		}
		id := r.Id
		if id == "" {
			s.logger.Warningf("Broken record, empty id: %s.", info.Name())
			continue
		}
		s.recordCache[id] = info.Name()
	}
	return nil
}

func (s *LocalStorage) SaveRecords(r *rpb.RecordSet) error {
	mergedSet := records.MergeRecordSets(
		ReadJSON[rpb.RecordSet](s.fs, filepath.Join(r.Id, recordsetFileName)).
			OrElse(&rpb.RecordSet{}), r)
	return s.writeRecordSet(mergedSet)
}

func (s *LocalStorage) writeRecordSet(r *rpb.RecordSet) error {
	root := r.Id
	if root == "" {
		return fmt.Errorf("Record must have an ID")
	}
	s.logger.Debugf("Saving record to %s", root)
	if err := s.fs.MkDir(root); err != nil {
		return err
	}
	if err := s.fs.WriteJSON(filepath.Join(root, recordsetFileName), r); err != nil {
		return err
	}
	for i, r := range r.Records {
		if r.Source != nil && r.Source.Url != "" {
			pageView := filepath.Join(root, fmt.Sprintf("page_view_record_%d", i))
			s.fs.MkDir(pageView)
			s.savePageView(r.Source.Url, pageView)
			r.Files = append(r.Files, &rpb.File{
				FileId:   "page_view_png",
				LocalUrl: pageView + "/page.png",
			}, &rpb.File{
				FileId:   "page_view_pdf",
				LocalUrl: pageView + "/page.pdf",
			})
		}

		for _, link := range r.Links {
			if util.IsYoutubeLink(link) && strings.Contains(link, "v=") {
				s.logger.Debugf("Found youtube link: %s", link)
				if err := util.DownloadYoutube(link, s.fs.Resolve(root)); err != nil {
					s.logger.Warningf("Failed to download youtube video: %s", err)
				}
			}

		}

		for _, file := range r.GetFiles() {
			if err := s.downloadFile(root, file); err != nil {
				continue
			}
		}

	}
	s.logger.Infof("Saved new record to %s", root)
	return nil
}

func (s *LocalStorage) GetFile(id string, filename string) optional.Optional[[]byte] {
	return s.fs.Read(filepath.Join(id, filename))
}

func (s *LocalStorage) ListRecords() optional.Optional[[]*rpb.RecordSet] {
	s.refreshCache()
	result := []*rpb.RecordSet{}
	for _, path := range s.recordCache {
		ReadJSON[rpb.RecordSet](s.fs, filepath.Join(path, recordsetFileName)).
			IfPresent(func(rs *rpb.RecordSet) {
				if rs.Id == "" {
					s.logger.Warningf("Broken record %s: empty id", path)
					return
				}
				result = append(result, rs)
			})
	}
	sort.Slice(result, func(i int, j int) bool {
		return result[i].Id < result[j].Id
	})
	return optional.Of(result)
}

func NewStorage(root string, webdriver firefox.WebDriver) Storage {
	log := util.NewLogger("storage")
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
		fs:          NewRelativeFS(root),
		downloader:  NewHttpDownloader(nil),
		recordCache: map[string]string{},
	}
}
