package chronist

import (
	"fmt"
	"strconv"
	"strings"
	"sort"

	"chronist/storage"
	"chronist/telegram"
	"chronist/util"

	rpb "chronist/proto/records"
)

type IChronist interface {
	FetchRequests() ([]*rpb.Record, error)
	SaveRequests(record []*rpb.Record) error
	SendStatusUpdate(source *rpb.Source, status rpb.FetchStatus, msg string) error
	GetCursor() int64
	SetCursor(cursor int64)
}

type Chronist struct {
	IChronist

	cursor  int64
	logger  *util.Logger
	tg      *telegram.Bot
	storage *storage.Storage
}

func NewChronist(cursor int64, tg *telegram.Bot, st *storage.Storage) IChronist {
	return &Chronist{
		cursor:  cursor,
		logger:  util.NewLogger("chronist"),
		storage: st,
		tg:      tg,
	}
}

func (c *Chronist) GetCursor() int64 {
	return c.cursor
}

func (c *Chronist) SetCursor(cursor int64) {
	c.cursor = cursor
}

func FromTelegramUpdate(upd *telegram.Update, baseRecord *rpb.Record) *rpb.Record {
	msg := upd.Message

	result := &rpb.Record{
		Source: &rpb.Source{
			SenderId:  fmt.Sprintf("%d", msg.From.ID),
			ChannelId: fmt.Sprintf("%d", msg.Chat.ID),
			MessageId: fmt.Sprintf("%d", msg.MessageID),
		},
		RecordId: fmt.Sprintf("%d", upd.UpdateID),
	}
	for _, e := range msg.Entities {
		if e.Type == "url" {
			result.Links = append(result.Links, e.URL)
		}
	}
	result.TextContent = strings.Replace(msg.Text, "\n\n", "\n", -1)
	if msg.Video != nil {
		result.Files = append(result.Files, &rpb.File{FileId: msg.Video.FileID})
	}
	if msg.Photo != nil {
		result.Files = append(result.Files, &rpb.File{
			FileId: telegram.GetLargestImage(msg.Photo).FileID,
		})
	}
	if baseRecord != nil {
		result.Links = append(result.Links, baseRecord.Links...)
		result.Files = append(result.Files, baseRecord.Files...)
		newText := result.TextContent
		if strings.Contains(baseRecord.TextContent, newText) {
			newText = baseRecord.TextContent
		} else if !strings.Contains(newText, baseRecord.TextContent) {
			newText += "\n" + baseRecord.TextContent
		}
		result.TextContent = newText
	}
	result.Links = util.Unique(append(result.Links, util.FindWebLinks(result.TextContent)...))
	sort.Strings(result.Links)
	return result
}

func (ch *Chronist) SaveRequests(requests []*rpb.Record) error {
	for _, req := range requests {
		ch.SendStatusUpdate(req.Source, rpb.FetchStatus_IN_PROGRESS, "")
		if err := ch.storage.SaveRecord(req); err != nil {
			ch.SendStatusUpdate(req.Source, rpb.FetchStatus_FAIL, err.Error())
		} else {
			ch.SendStatusUpdate(req.Source, rpb.FetchStatus_SUCCESS, "")
		}
	}
	return nil
}

func (ch *Chronist) SendStatusUpdate(source *rpb.Source, status rpb.FetchStatus, msg string) error {
	sender, _ := strconv.Atoi(source.SenderId)
	message, _ := strconv.Atoi(source.MessageId)
	ch.logger.Infof("Fetch status is %v", status)
	if status == rpb.FetchStatus_FAIL || status == rpb.FetchStatus_SUCCESS {
		ch.tg.SendMessage(int64(sender), int64(message), status.String())
	}
	return nil
}

func (ch *Chronist) FetchRequests() ([]*rpb.Record, error) {
	records := map[string]*rpb.Record{}
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
			records[key] = FromTelegramUpdate(upd, records[key])
		}
		ch.logger.Infof("Loaded %d updates into %d records", len(updates), len(records))
	}
	for _, record := range records {
		if len(record.Files) == 0 {
			continue
		}
		for _, file := range record.Files {
			fileURL, err := ch.tg.GetFile(file.FileId)
			if err != nil {
				ch.logger.Errorf("Cannot get actual file url for %s: %s", file.FileId, err)
				continue
			}
			file.FileUrl = ch.tg.GetUrl(fileURL)
		}
	}
	return util.Values(records), nil
}
