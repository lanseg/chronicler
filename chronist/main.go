package main

import (
    "fmt"
    "os"
    "log"
    "strconv"
    "strings"

    "chronist/util"
    "chronist/telegram"
    "chronist/storage"
)

const (
  privateChatId = int64(0)
)

var (
  logger = log.Default()
)

type IChronist interface {
  
  FetchRequests() ([]*storage.Record, error)
  StoreRequest(record *storage.Record) error
}

type Chronist struct {
  IChronist
  
  cursor int64
  logger *log.Logger
  tg *telegram.TelegramBot
}

func (ch *Chronist) FetchRequests() ([]*storage.Record, error) {
  records := map[string]*storage.Record{}    
  var updates []*telegram.Update = nil
  
  for len(updates) == 0 {
    ch.logger.Printf("Loading all updates starting from %d", ch.cursor)
    updates, _ = ch.tg.GetUpdates(privateChatId, ch.cursor, 100, 100, []string{})
    for _, upd := range updates {
      if ch.cursor < upd.UpdateId {
        ch.cursor = upd.UpdateId
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

func FromTelegramUpdate(upd *telegram.Update) *storage.Record {
    msg := upd.Message  
    result := &storage.Record{
      Source  : &storage.Source {
        SenderId: fmt.Sprintf("%d", msg.From.Id),
        ChannelId: fmt.Sprintf("%d", msg.Chat.Id),
        MessageId: fmt.Sprintf("%d", msg.MessageId),
      },
      RecordId: fmt.Sprintf("%d", upd.UpdateId),
      Links:    []string{},
    }
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

func getCursor() int64 {
  bytes, _ := os.ReadFile("cursor.txt")
  num, _ := strconv.Atoi(string(bytes))
  return int64(num)
}

func saveCursor(cursor int64) {
  os.WriteFile("cursor.txt", []byte(fmt.Sprintf("%d", cursor)), 0644)
}
  
func main() {
    tgApiKey := os.Args[1]
    storageRoot := "chronist_storage"
    stg := storage.NewStorage(storageRoot)
    chr := &Chronist {
      cursor: getCursor(),
      logger: log.Default(),
      tg: telegram.NewBot(tgApiKey),
    }

    newRequests, err := chr.FetchRequests()
    if err != nil {
      fmt.Println(err.Error())
      return
    }
    requestBySource := util.GroupBy(newRequests, func (r *storage.Record) *storage.Source {
      return r.Source
    })
    for src, reqs := range requestBySource {
      success := []*storage.Record{}
      failure := []*storage.Record{}
      for _, req := range reqs {
        if err := stg.SaveRecord(req); err != nil {
          failure = append(failure, req);
          logger.Printf("ERROR: failed to save record %v: %s\n", req, err)
        } else {
          success = append(success, req)
          logger.Printf("Saved record %v\n", req)
        }
      }      
      id, _ := strconv.Atoi(src.ChannelId)
      chr.tg.SendMessage(int64(id),
                         fmt.Sprintf("Saved %d new records, failed to save: %d", 
                                     len(success), len(failure)))
    }
    saveCursor(chr.cursor + 1)
}
