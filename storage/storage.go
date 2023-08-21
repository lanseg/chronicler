package storage

import (
	"net/url"
	"path"
	"path/filepath"
	"sort"
	"strings"

	"chronicler/records"
	rpb "chronicler/records/proto"
	"chronicler/util"

	"github.com/lanseg/golang-commons/optional"
)

const (
	recordsetFileName = "record.json"
)

type Storage interface {
	SaveRecords(r *rpb.RecordSet) error
	ListRecords() optional.Optional[[]*rpb.RecordSet]
	GetFile(id string, filename string) optional.Optional[[]byte]
}

type LocalStorage struct {
	Storage

	downloader *HttpDownloader
	fs         *RelativeFS
	runner     *util.Runner

	recordCache map[string]string
	logger      *util.Logger
	root        string
}

func (s *LocalStorage) refreshCache() error {
	s.logger.Debugf("Refreshing record cache")
	s.recordCache = map[string]string{}
	files, err := s.fs.ListFiles("")
	if err != nil {
		return err
	}
	for _, info := range files {
		ReadJSON[rpb.RecordSet](s.fs, filepath.Join(info.Name(), recordsetFileName)).
			IfPresent(func(record *rpb.RecordSet) {
				id := record.Id
				if id == "" {
					id = records.GetRecordSetId(record)
				}
				s.recordCache[id] = info.Name()
			})
	}
	return nil
}

func (s *LocalStorage) writeRecordSet(r *rpb.RecordSet) error {
	root := records.GetRecordSetId(r)
	s.logger.Debugf("Saving record to %s", root)
	if err := s.fs.MkDir(root); err != nil {
		return err
	}
	if err := s.fs.WriteJSON(filepath.Join(root, recordsetFileName), r); err != nil {
		return err
	}
	for _, r := range r.Records {
		for _, link := range r.Links {
			if util.IsYoutubeLink(link) && strings.Contains(link, "v=") {
				s.logger.Debugf("Found youtube link: %s", link)
				if err := util.DownloadYoutube(link, s.fs.Resolve(root)); err != nil {
					s.logger.Warningf("Failed to download youtube video: %s", err)
				}
			}
		}

		for _, file := range r.GetFiles() {
			fileUrl, err := url.Parse(file.GetFileUrl())
			if err != nil || fileUrl.String() == "" {
				s.logger.Warningf("Malformed url for file: %s", file)
				continue
			}
			fname := path.Base(fileUrl.Path)
			if err := s.downloader.Download(file.GetFileUrl(), s.fs.Resolve(filepath.Join(root, fname))); err != nil {
				s.logger.Warningf("Failed to download file: %s: %s", file, err)
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
					rs.Id = records.GetRecordSetId(rs)
				}
				result = append(result, rs)
			})
	}
	sort.Slice(result, func(i int, j int) bool {
		return result[i].Request.String() < result[j].Request.String()
	})
	return optional.Of(result)
}

func (s *LocalStorage) SaveRecords(r *rpb.RecordSet) error {
	id := records.GetRecordSetId(r)
	existing := ReadJSON[rpb.RecordSet](s.fs, filepath.Join(id, recordsetFileName)).
		OrElse(&rpb.RecordSet{})
	return s.writeRecordSet(records.MergeRecordSets(existing, r))
}

func NewStorage(root string) Storage {
	log := util.NewLogger("storage")
	log.Infof("Storage root set to \"%s\"", root)
	return &LocalStorage{
		root:        root,
		logger:      log,
		runner:      util.NewRunner(),
		fs:          NewRelativeFS(root),
		downloader:  NewHttpDownloader(nil),
		recordCache: map[string]string{},
	}
}
