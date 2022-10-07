package main

import (
    "flag"
	"fmt"
	"os"
	"strconv"
	"strings"

	"chronist/storage"
	"chronist/telegram"
	"chronist/util"
)

const (
	privateChatId = int64(0)
    tgBotKeyFlag = "telegram_bot_key"
    storageRootFlag = "storage_root"
)

var (
  telegramBotKey = flag.String(tgBotKeyFlag, "", "A key for the telegram bot api.")
  storageRoot = flag.String(storageRootFlag, "chronist_storage", "A local folder to save downloads.")
)

type IChronist interface {
	FetchRequests() ([]*storage.Record, error)
	StoreRequest(record *storage.Record) error
}

type Chronist struct {
	IChronist

	cursor int64
	logger *util.Logger
	tg     *telegram.Bot
}

func (ch *Chronist) FetchRequests() ([]*storage.Record, error) {
	records := map[string]*storage.Record{}
	var updates []*telegram.Update = nil

	for len(updates) == 0 {
		ch.logger.Infof("Loading all updates starting from %d", ch.cursor)
		updates, _ = ch.tg.GetUpdates(privateChatId, ch.cursor, 100, 100, []string{})
		for _, upd := range updates {
			if ch.cursor < upd.UpdateID {
				ch.cursor = upd.UpdateID
			}
			if upd.Message == nil {
				continue
			}
			msg := upd.Message
			key := fmt.Sprintf("%d_%d_%d", msg.Chat.ID, msg.From.ID, msg.Date)
			newRecord := FromTelegramUpdate(upd)

			if oldRecord, ok := records[key]; ok {
				oldRecord.Merge(newRecord)
			} else {
				records[key] = newRecord
			}
		}
		ch.logger.Infof("Loaded %d updates into %d records", len(updates), len(records))
	}
	for _, record := range records {
		if len(record.Files) == 0 {
			continue
		}
		for _, file := range record.Files {
			fileURL, err := ch.tg.GetFile(file.FileID)
			if err != nil {
				ch.logger.Errorf("Cannot get actual file url for %s: %s", file.FileID, err)
				continue
			}
			file.FileURL = ch.tg.GetUrl(fileURL)
		}
	}
	return util.Values(records), nil
}

func FromTelegramUpdate(upd *telegram.Update) *storage.Record {
	msg := upd.Message
	result := &storage.Record{
		Source: &storage.Source{
			SenderID:  fmt.Sprintf("%d", msg.From.ID),
			ChannelID: fmt.Sprintf("%d", msg.Chat.ID),
			MessageID: fmt.Sprintf("%d", msg.MessageID),
		},
		RecordID: fmt.Sprintf("%d", upd.UpdateID),
		Links:    []string{},
	}
	for _, e := range msg.Entities {
		if e.Type == "url" {
			result.Links = append(result.Links, e.URL)
		}
	}
	result.TextContent = strings.Replace(msg.Text, "\n\n", "\n", -1)
	if msg.Video != nil {
		result.AddFile(msg.Video.FileID)
	}
	if msg.Photo != nil {
		result.AddFile(telegram.GetLargestImage(msg.Photo).FileID)
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
    flag.Parse()  
    logger := util.NewLogger("main")  

    if len(*telegramBotKey) == 0{
      logger.Errorf("No telegram bot key defined, please set it with --%s=\"...\"", tgBotKeyFlag)
      return
    }

	stg := storage.NewStorage(*storageRoot)
	chr := &Chronist{
		cursor: getCursor(),
		logger: util.NewLogger("chronist"),
		tg:     telegram.NewBot(*telegramBotKey),
	}

	newRequests, err := chr.FetchRequests()
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	requestBySource := util.GroupBy(newRequests, func(r *storage.Record) *storage.Source {
		return r.Source
	})
	for src, reqs := range requestBySource {
		success := []*storage.Record{}
		failure := []*storage.Record{}
		for _, req := range reqs {
			if err := stg.SaveRecord(req); err != nil {
				failure = append(failure, req)
				logger.Errorf("failed to save record %v: %s\n", req, err)
			} else {
				success = append(success, req)
				logger.Infof("Saved record %v\n", req)
			}
		}
		id, _ := strconv.Atoi(src.ChannelID)
		chr.tg.SendMessage(int64(id),
			fmt.Sprintf("Saved %d new records, failed to save: %d",
				len(success), len(failure)))
	}
	saveCursor(chr.cursor + 1)
}
