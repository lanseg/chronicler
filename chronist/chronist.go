package chronist

import (
	"fmt"
	"strings"

	"chronist/storage"
	"chronist/telegram"
	"chronist/util"
)

type IChronist interface {
	FetchRequests() ([]*storage.Record, error)
	StoreRequest(record *storage.Record) error
	GetCursor() int64
}

type Chronist struct {
	IChronist

	cursor int64
	logger *util.Logger
	tg     *telegram.Bot
}

func NewChronist(cursor int64, tg *telegram.Bot) IChronist {
	return &Chronist {
		cursor: cursor,
		logger: util.NewLogger("chronist"),
		tg: tg,
	}
}

func (c *Chronist) GetCursor() int64 {
	return c.cursor
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

func (ch *Chronist) FetchRequests() ([]*storage.Record, error) {
	records := map[string]*storage.Record{}
	var updates []*telegram.Update = nil

	for len(updates) == 0 {
		ch.logger.Infof("Loading all updates starting from %d", ch.cursor)
		updates, _ = ch.tg.GetUpdates(int64(0), ch.cursor, 100, 100, []string{})
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
