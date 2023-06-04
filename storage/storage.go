package storage

import (
	"crypto/sha512"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"

	rpb "chronicler/proto/records"
	"chronicler/util"
)

const (
	userAgent = "curl/7.54"
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
}

type LocalStorage struct {
	Storage

	httpClient *http.Client
	logger     *util.Logger
	root       string
}

func (s *LocalStorage) path(relativePath string) string {
	return filepath.Join(s.root, relativePath)
}

func (s *LocalStorage) mkdir(path string) error {
	recordRoot := s.path(path)
	s.logger.Infof("Creating directory at [%s]/%s: %s", s.root, path, recordRoot)
	if err := os.MkdirAll(recordRoot, os.ModePerm); err != nil {
		return err
	}
	return nil
}

func (s *LocalStorage) writeFile(path string, value []byte) error {
	return os.WriteFile(s.path(path), value, os.ModePerm)
}

func (s *LocalStorage) copyReader(src io.ReadCloser, dst string) error {
	defer src.Close()

	targetFile, err := os.Create(s.path(dst))
	defer targetFile.Close()
	if err != nil {
		return err
	}
	_, err = io.Copy(targetFile, src)
	if err != nil {
		return err
	}
	return nil
}

func (s *LocalStorage) get(link string) (*http.Response, error) {
	req, err := http.NewRequest("GET", link, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", userAgent)
	return s.httpClient.Do(req)
}

func (s *LocalStorage) downloadURL(url string, target string) error {
	s.logger.Debugf("Downloading file from %s to %s", url, target)
	resp, err := s.get(url)
	if err != nil {
		return err
	}
	return s.copyReader(resp.Body, target)
}

func (s *LocalStorage) ListRecords() ([]*rpb.RecordSet, error) {
	result := []*rpb.RecordSet{}
	filepath.Walk(s.root,
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
			result = append(result, rs)
			return nil
		})
	sort.Slice(result, func(i int, j int) bool {
		return result[i].String() < result[j].String()
	})
	return result, nil
}

func (s *LocalStorage) SaveRecords(r *rpb.RecordSet) error {
	root := getRecordSetId(r)
	s.logger.Debugf("Saving record to %s", s.path(root))
	if err := s.mkdir(root); err != nil {
		return err
	}
	bytes, err := json.Marshal(r)
	if err != nil {
		return fmt.Errorf("Json marshalling error: %s", err)
	}
	if err := s.writeFile(filepath.Join(root, "record.json"), bytes); err != nil {
		return err
	}
	for _, r := range r.Records {
		for _, link := range r.Links {
			if util.IsYoutubeLink(link) && strings.Contains(link, "v=") {
				s.logger.Debugf("Found youtube link: %s", link)
				if err := util.DownloadYoutube(link, s.path(root)); err != nil {
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
			if err := s.downloadURL(file.GetFileUrl(), filepath.Join(root, fname)); err != nil {
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
		root:       root,
		httpClient: &http.Client{},
		logger:     log,
	}
}
