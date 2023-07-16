package storage

import (
	"crypto/sha512"
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"

	rpb "chronicler/proto/records"
	"chronicler/util"
)

func getRecordSetId(set *rpb.RecordSet) string {
	if set.Id != "" {
		return set.Id
	}
	checksum := []byte{}
	if set.Request != nil {
		checksum = append(checksum, []byte(set.Request.Source.SenderId)...)
		checksum = append(checksum, []byte(set.Request.Source.ChannelId)...)
		checksum = append(checksum, []byte(set.Request.Source.MessageId)...)
		checksum = append(checksum, []byte(set.Request.Source.Url)...)
		checksum = append(checksum, byte(set.Request.Source.Type))
	}
	return fmt.Sprintf("%x", sha512.Sum512(checksum))
}

type Storage interface {
	SaveRecords(r *rpb.RecordSet) error
	ListRecords() ([]*rpb.RecordSet, error)
	GetFile(id string, filename string) ([]byte, error)
}

type LocalStorage struct {
	Storage

	downloader *HttpDownloader
	fs         *RelativeFS

	recordCache map[string]string
	logger      *util.Logger
	root        string
}

func (s *LocalStorage) refreshCache() error {
	s.logger.Debugf("Refreshing record cache")
	s.recordCache = map[string]string{}
	return filepath.Walk(s.root,
		func(path string, info os.FileInfo, err error) error {
			if filepath.Base(path) != "record.json" {
				return nil
			}
			b, err := os.ReadFile(path)
			if err != nil {
				s.logger.Warningf("Error reading file: %s", err)
				return err
			}
			rs := &rpb.RecordSet{}
			if err = json.Unmarshal(b, &rs); err != nil {
				s.logger.Warningf("Error unmarshalling file: %s", err)
				return err
			}
			id := rs.Id
			if rs.Id == "" {
				id = getRecordSetId(rs)
			}
			s.recordCache[id] = filepath.Dir(path)
			return nil
		})
}

func (s *LocalStorage) GetFile(id string, filename string) ([]byte, error) {
	return s.fs.Read(filepath.Join(id, filename))
}

func (s *LocalStorage) ListRecords() ([]*rpb.RecordSet, error) {
	s.refreshCache()
	result := []*rpb.RecordSet{}
	for _, path := range s.recordCache {
		b, err := os.ReadFile(filepath.Join(path, "record.json"))
		if err != nil {
			s.logger.Warningf("Error reading file: %s", err)
			continue
		}
		rs := &rpb.RecordSet{}
		if err = json.Unmarshal(b, &rs); err != nil {
			s.logger.Warningf("Error unmarshalling file: %s", err)
			continue
		}
		if rs.Id == "" {
			rs.Id = getRecordSetId(rs)
		}
		result = append(result, rs)
	}
	sort.Slice(result, func(i int, j int) bool {
		return result[i].Request.String() < result[j].Request.String()
	})
	return result, nil
}

func (s *LocalStorage) SaveRecords(r *rpb.RecordSet) error {
	root := getRecordSetId(r)
	s.logger.Debugf("Saving record to %s", root)
	if err := s.fs.MkDir(root); err != nil {
		return err
	}
	if err := s.fs.WriteJSON(filepath.Join(root, "record.json"), r); err != nil {
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

func NewStorage(root string) Storage {
	log := util.NewLogger("storage")
	log.Infof("Storage root set to \"%s\"", root)
	return &LocalStorage{
		root:        root,
		logger:      log,
		fs:          NewRelativeFS(root),
		downloader:  NewHttpDownloader(nil),
		recordCache: map[string]string{},
	}
}
