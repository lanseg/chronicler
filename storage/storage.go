package storage

import (
	"fmt"
	"io"
	"os"
	"strings"
	"net/http"
	"path/filepath"
    
    "chronist/util"
)

type File struct {
	FileID  string
	FileURL string
}

type Source struct {
	SenderID  string
	ChannelID string
	MessageID string
}

type Record struct {
	RecordID    string
	Source      *Source
	Files       []*File
	Links       []string
	TextContent string
}

func (r *Record) String() string {
	return fmt.Sprintf(
		"{RecordID: %s, source: %v, files: %v, links: %v, TextContent: %s}",
		r.RecordID, r.Source, r.Files, r.Links, r.TextContent)
}

func (r *Record) AddFile(fileId string) {
	if r.Files == nil {
		r.Files = []*File{}
	}

	r.Files = append(r.Files, &File{FileID: fileId, FileURL: ""})
}

func (r *Record) Merge(other *Record) {
	newFiles := map[string]*File{}
	for _, f := range r.Files {
		newFiles[f.FileID] = f
	}
	
	for _, f := range other.Files {
		newFiles[f.FileID] = f
	}

	newLinks := map[string]bool{}
	for _, l := range r.Links {
		newLinks[l] = true
	}

	for _, l := range other.Links {
		newLinks[l] = true
	}

	newText := r.TextContent
	if strings.Contains(other.TextContent, newText) {
		newText = other.TextContent
	} else if !strings.Contains(newText, other.TextContent) {
		newText += "\n" + other.TextContent
	}

	r.Files = util.Values(newFiles)
	r.Links = util.Keys(newLinks)
	r.TextContent = newText
}

type IStorage interface {
	SaveRecord(r *Record) error
}

type Storage struct {
	IStorage

	httpClient *http.Client
	logger     *util.Logger
	root       string
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

func (s *Storage) SaveRecord(r *Record) error {
	if err := os.MkdirAll(filepath.Join(s.root, r.RecordID), os.ModePerm); err != nil {
		return err
	}

	if len(r.Links) > 0 {
		if err := s.saveLines(filepath.Join(r.RecordID, "links.txt"), r.Links); err != nil {
			return err
		}
		for _, link := range r.Links {
          if !util.IsYoutubeLink(link) {
            continue
          }
          if err := util.DownloadYoutube(link, filepath.Join(s.root, r.RecordID)); err != nil {
            s.logger.Warningf("Failed to download youtube video: %s", err)
          }
        }
	}

	if len(r.TextContent) > 0 {
		if err := s.saveLines(filepath.Join(r.RecordID, "text.txt"), []string{
			r.TextContent,
		}); err != nil {
			return err
		}
	}
	
	if len(r.Files) > 0 {
		for i, file := range r.Files {
			fname := fmt.Sprintf("file_%d", i)
			if err := s.downloadURL(file.FileURL, filepath.Join(r.RecordID, fname)); err != nil {
				return err
			}
		}
	}

	s.logger.Infof("Saved new record to %s", filepath.Join(s.root, r.RecordID)) 
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
