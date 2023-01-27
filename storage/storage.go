package storage

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	rpb "chronicler/proto/records"
	"chronicler/util"
)

type IStorage interface {
	SaveRecords(name string, r *rpb.RecordSet) error
}

type Storage struct {
	IStorage

	httpClient *http.Client
	logger     *util.Logger
	root       string
}

func (s *Storage) saveText(name string, text string) error {
	return os.WriteFile(filepath.Join(s.root, name), []byte(text), os.ModePerm)
}

func (s *Storage) saveLines(name string, lines []string) error {
	return os.WriteFile(
		filepath.Join(s.root, name),
		[]byte(strings.Join(lines, "\n")),
		os.ModePerm)
}

func (s *Storage) downloadURL(url string, target string) error {
	resp, err := s.httpClient.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	targetFile, err := os.Create(filepath.Join(s.root, target))
	if err != nil {
		return err
	}
	defer targetFile.Close()

	_, err = io.Copy(targetFile, resp.Body)
	if err != nil {
		return err
	}

	return nil
}

func (s *Storage) SaveRecords(recordRoot string, r *rpb.RecordSet) error {
	if err := os.MkdirAll(filepath.Join(s.root, recordRoot), os.ModePerm); err != nil {
		return err
	}
	bytes, err := json.Marshal(r)
	if err != nil {
		return fmt.Errorf("Json marshalling error: %s", err)
	}
	if err := s.saveText(filepath.Join(recordRoot, "record.json"), string(bytes)); err != nil {
		return err
	}
	for _, r := range r.Records {
		if len(r.Links) > 0 {
			for _, link := range r.Links {
				if util.IsYoutubeLink(link) {
					if err := util.DownloadYoutube(link, filepath.Join(s.root, recordRoot)); err != nil {
						s.logger.Warningf("Failed to download youtube video: %s", err)
					}
				}
			}
		}

		if len(r.Files) > 0 {
			for i, file := range r.GetFiles() {
				if file.GetFileUrl() == "" {
					s.logger.Warningf("Record with source %s has a without an url: %s",
						r.Source, file)
					continue
				}
				fileUrl := file.GetFileUrl()
				fnamePos := strings.LastIndex(fileUrl, "/")
				fname := fmt.Sprintf("%d_%s", i, fileUrl[fnamePos+1:])
				if err := s.downloadURL(file.GetFileUrl(), filepath.Join(recordRoot, fname)); err != nil {
					s.logger.Warningf("Failed to download file: %s: %s", file, err)
				}
			}
		}

	}
	s.logger.Infof("Saved new record to %s", recordRoot)
	return nil
}

func NewStorage(root string) *Storage {
	log := util.NewLogger("storage")
	log.Infof("Storage root set to %s", root)
	return &Storage{
		root:       root,
		httpClient: &http.Client{},
		logger:     log,
	}
}
