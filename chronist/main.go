package main

import (
    "fmt"
    "os"
    "log"
    "strings"
    "chronist/telegram"
    "path/filepath"
)

type RecordType int

const (
  UNKNOWN RecordType = 0
  TEXT    RecordType = 1
  VIDEO   RecordType = 2
  IMAGE   RecordType = 3
  TWITTER RecordType = 4
)

type Record struct {
  recordId    string
  recordType  RecordType
  fileId      string
  links       []string
  textContent string
}

func readUpdate(upd *telegram.Update) *Record {
    if upd.Message == nil {
      return nil
    }  
    result := &Record{
      recordId: fmt.Sprintf("%d", upd.UpdateId),
      links:    []string{},
    }
    msg := upd.Message
    for _, e := range msg.Entities {
      if e.Type == "url" {
        result.links = append(result.links, e.Url)
      }
    }
    result.textContent = strings.Replace(msg.Text, "\n\n", "\n", -1)
    if msg.Video != nil {
      result.recordType = VIDEO
      result.fileId = msg.Video.FileId
    }
    if msg.Photo != nil {
      result.recordType = IMAGE
      result.fileId = telegram.GetLargestImage(msg.Photo).FileId
    }
    return result
}

type IChronist interface {
  
  FetchRequests() ([]*Record, error);
}

type Chronist struct {
  IChronist
  
  storageRoot string
  logger *log.Logger
  tg *telegram.TelegramBot
}

func (ch *Chronist) FetchRequests() ([]*Record, error) {
  updId := int64(0)
  records := []*Record{}    
  var updates []*telegram.Update = nil
  
  for len(updates) == 0 {
    ch.logger.Printf("Loading all updates starting from %d", updId)
    updates, _ = ch.tg.GetUpdates(updId, 100, 100, []string{})
    for _, upd := range updates {
      if updId < upd.UpdateId {
        updId = upd.UpdateId
      }
      records = append(records, readUpdate(upd))
    }
    ch.logger.Printf("Loaded %d updates into %d records", len(updates), len(records))
  }
  return records, nil
}

func (ch *Chronist) StoreRequest(record *Record) error {
  recordRoot := filepath.Join(ch.storageRoot, record.recordId)
  if err := os.MkdirAll(recordRoot, os.ModePerm); err != nil {
    return err
  }
  if record.recordType == VIDEO || record.recordType == IMAGE {
    file, err := ch.tg.GetFile(record.fileId)
    if err != nil {
      return err
    }
    targetName := "video"
    if record.recordType == IMAGE {
      targetName = "image"
    }
    return ch.tg.Download(file, filepath.Join(recordRoot, targetName))
  }

  f, err := os.Create(filepath.Join(recordRoot, "textdata.txt"))
  if err != nil {
    return err
  }
  defer f.Close()
  _, err = f.WriteString(record.textContent)
  if err != nil {
    return err
  }
  
  if len(record.links) == 0 {
    return nil
  }
  f, err = os.Create(filepath.Join(recordRoot, "links.txt"))
  if err != nil {
    return err
  }
  defer f.Close()  
  _, err = f.WriteString(strings.Join(record.links, "\n"))
  if err != nil {
    return err
  }
  return nil
}

func NewChronist(tg *telegram.TelegramBot) *Chronist {
  return &Chronist {
    storageRoot: "chronist_storage",
    logger: log.Default(),
    tg: tg,
  }
}

func main() {
    chr := NewChronist(telegram.NewBot(os.Args[1]))
    reqs, _ := chr.FetchRequests()
    for _, req := range reqs {
      if err := chr.StoreRequest(req); err != nil {
        fmt.Printf("ERROR: %s", err)
      }
    }
}
