package storage

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	rpb "chronist/proto/records"
	"chronist/util"

	"github.com/golang/protobuf/proto"
)

type IStorage interface {
	SaveRecord(r *rpb.Record) error
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

func (s *Storage) SaveRecord(r *rpb.Record) error {
	var recordRoot string
	if r.GetParentRecordId() != "" {
		recordRoot = filepath.Join(r.GetParentRecordId(), r.GetRecordId())
	} else {
		recordRoot = filepath.Join(r.GetRecordId())
	}
	if err := os.MkdirAll(filepath.Join(s.root, recordRoot), os.ModePerm); err != nil {
		return err
	}

	if len(r.Links) > 0 {
		if err := s.saveLines(filepath.Join(recordRoot, "links.txt"), r.Links); err != nil {
			return err
		}
		for _, link := range r.Links {
			if util.IsYoutubeLink(link) {
				if err := util.DownloadYoutube(link, filepath.Join(s.root, recordRoot)); err != nil {
					s.logger.Warningf("Failed to download youtube video: %s", err)
				}
			}
		}
	}

	if len(r.TextContent) > 0 {
		if err := s.saveText(filepath.Join(recordRoot, "text.txt"), r.TextContent); err != nil {
			return err
		}
	}

	if len(r.Files) > 0 {
		for i, file := range r.GetFiles() {
			fname := fmt.Sprintf("file_%d", i)
			if err := s.downloadURL(file.GetFileUrl(), filepath.Join(recordRoot, fname)); err != nil {
				return err
			}
		}
	}

	if err := s.saveText(filepath.Join(recordRoot, ".metadata"), proto.MarshalTextString(r)); err != nil {
		return err
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
