package main

import (
    "fmt"
    "os"
    "log"
    "strings"

    "chronist/util"
    "chronist/telegram"
    "chronist/storage"
)

const (
  privateChatId = int64(-1480532340)
)

type IChronist interface {
  
  FetchRequests() ([]*storage.Record, error)
  StoreRequest(record *storage.Record) error
}

type Chronist struct {
  IChronist
  
  logger *log.Logger
  tg *telegram.TelegramBot
}

func (ch *Chronist) FetchRequests() ([]*storage.Record, error) {
  updId := int64(0)
  records := map[string]*storage.Record{}    
  var updates []*telegram.Update = nil
  
  for len(updates) == 0 {
    ch.logger.Printf("Loading all updates starting from %d", updId)
    updates, _ = ch.tg.GetUpdates(privateChatId, updId, 100, 100, []string{})
    for _, upd := range updates {
      if updId < upd.UpdateId {
        updId = upd.UpdateId
      }
      if upd.Message == nil {
        continue
      }  
      msg := upd.Message
      key := fmt.Sprintf("%d_%d_%d", msg.Chat.Id, msg.From.Id, msg.Date)
      newRecord := FromTelegramUpdate(upd)
      
      if oldRecord, ok := records[key]; ok {
        oldRecord.Merge(newRecord)
      } else {
        records[key] = newRecord
      }
    }
    ch.logger.Printf("Loaded %d updates into %d records", len(updates), len(records))
  }
  for _, record := range records {
    if len(record.Files) == 0 {
      continue
    }
    for _, file := range record.Files {
      fileUrl, err := ch.tg.GetFile(file.FileId)
      if err != nil {
        ch.logger.Printf("Cannot get actual file url for %s: %s\n", file.FileId, err)
        continue
      }
      file.FileUrl = ch.tg.GetUrl(fileUrl)
    }
  }
  return util.Values(records), nil
}

func NewChronist(telegramKey string) *Chronist {
  return &Chronist {
    logger: log.Default(),
    tg: telegram.NewBot(telegramKey),
  }
}

func FromTelegramUpdate(upd *telegram.Update) *storage.Record {
    result := &storage.Record{
      RecordId: fmt.Sprintf("%d", upd.UpdateId),
      Links:    []string{},
    }
    msg := upd.Message
    for _, e := range msg.Entities {
      if e.Type == "url" {
        result.Links = append(result.Links, e.Url)
      }
    }
    result.TextContent = strings.Replace(msg.Text, "\n\n", "\n", -1)
    if msg.Video != nil {
      result.AddFile(msg.Video.FileId)
    }
    if msg.Photo != nil {
      result.AddFile(telegram.GetLargestImage(msg.Photo).FileId)
    }
    return result
}

func main() {
    chr := NewChronist(os.Args[1])
    storage := storage.NewStorage("chronist_storage")

    reqs, _ := chr.FetchRequests()
    for _, req := range reqs {
      if err := storage.SaveRecord(req); err != nil {
        chr.tg.SendMessage(privateChatId, fmt.Sprintf("ERROR: %s", err))
        fmt.Printf("ERROR: %s", err)
      }
    }
}
