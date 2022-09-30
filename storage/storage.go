package storage

import (
    "io"
    "log"
    "os"
    "strings"
    
    "net/http"    
    "path/filepath"
)

type Record struct {
  RecordId    string
  FileId      string
  FileUrl     string
  Links       []string
  TextContent string
}

type IStorage interface {
    SaveRecord(r *Record) error
}

type Storage struct {
    IStorage
    
    httpClient *http.Client
    logger *log.Logger    
    root string
}

func (s *Storage) saveLines(name string, lines []string) error {
    return os.WriteFile(
        filepath.Join(s.root, name),
        []byte(strings.Join(lines, "\n")),
        os.ModePerm)
}

func (s *Storage) downloadUrl(url string, target string) error {
    resp, err := s.httpClient.Get(url)
    if err != nil {
        return err 
    }
    defer resp.Body.Close();    
   
    targetFile, err := os.Create(filepath.Join(s.root, target))
    if err != nil {
        return err
    }
    defer targetFile.Close();
    
    _, err = io.Copy(targetFile, resp.Body)
    if err != nil {
        return err
    }
    return nil
}

func (s *Storage) SaveRecord(r *Record) error {
  if err := os.MkdirAll(filepath.Join(s.root, r.RecordId), os.ModePerm); err != nil {
    return err
  }
  if len(r.Links) > 0 {
    if err := s.saveLines(filepath.Join(r.RecordId, "links.txt"), r.Links); err != nil {
      return err
    }
  }
  if len(r.TextContent) > 0 {
    if err := s.saveLines(filepath.Join(r.RecordId, "text.txt"), []string{
        r.TextContent,
    }); err != nil {
      return err
    }
  }
  if len(r.FileUrl) > 0 {
      if err := s.downloadUrl(r.FileUrl, filepath.Join(r.RecordId, "file")); err != nil {
          return err
      }
  }
  return nil
}

func NewStorage(root string) *Storage {
    return &Storage {
        root: root,
        httpClient: &http.Client{},
        logger: log.Default(),
    }
}
