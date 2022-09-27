package main

import (
    "fmt"
    "os"
    "strings"
    "chronist/telegram"
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
  recordType RecordType
  
  fileId      string
  links       []string
  textContent string
}

func (r *Record) toString() string {
  text := r.textContent
  if len(text) > 10 {
    text = text[:10] + "..."
  }
  return fmt.Sprintf("Record {type: %d, links: %s, file:%s, text: \"%s\"}",
                     r.recordType, r.links, r.fileId, r.textContent)
}


func readUpdate(upd *telegram.Update) *Record {
    if upd.Message == nil {
      return nil
    }  
    result := &Record{
      links: []string{},
    }
    msg := upd.Message
    for _, e := range msg.Entities {
      if e.Type == "text_link" {
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

func main() {
    b := telegram.NewBot(os.Args[1])
    updId := int64(0)
    var updates []*telegram.Update = nil
    
    for updates == nil || len(updates) == 0 {
      updates, _ = b.GetUpdates(updId, 100, 100, []string{})
      for _, upd := range updates {
        if updId < upd.UpdateId {
          updId = upd.UpdateId
        }
        fmt.Println(readUpdate(upd).toString())
      }
    }

    fmt.Printf("Total updates: %d\n", len(updates))
}
