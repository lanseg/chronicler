package chronicler

import (
	"fmt"
	"sort"
	"strings"

	"chronicler/telegram"
	"chronicler/util"

	"github.com/lanseg/golang-commons/collections"
	"github.com/lanseg/golang-commons/optional"

	rpb "chronicler/proto/records"
)

type TelegramSource struct {
	RequestRecordSource

	log      *util.Logger
	bot      *telegram.Bot
	cursor   int64
	records  chan *rpb.RecordSet
	requests chan *rpb.Request
}

func NewTelegramSource(bot *telegram.Bot) RequestRecordSource {
	src := &TelegramSource{
		log:      util.NewLogger("TelegramSource"),
		bot:      bot,
		cursor:   0,
		records:  make(chan *rpb.RecordSet),
		requests: make(chan *rpb.Request),
	}
	go src.fetchLoop()
	return src
}

func (ts *TelegramSource) groupRequests(updates []*telegram.Update) []*rpb.Request {
	return []*rpb.Request{}
}

func (ts *TelegramSource) groupRecords(updates []*telegram.Update) *rpb.RecordSet {
	recordByMedia := map[string][]*rpb.Record{}
	userById := map[string]*rpb.UserMetadata{}

	for _, upd := range updates {
		record, users := updateToRecords(upd)
		grpId := upd.Message.MediaGroupID
		if _, ok := recordByMedia[grpId]; !ok {
			recordByMedia[grpId] = []*rpb.Record{}
		}
		recordByMedia[grpId] = append(recordByMedia[grpId], record)
		for _, u := range users {
			userById[u.Id] = u
		}
	}

	result := &rpb.RecordSet{
		UserMetadata: collections.Values(userById),
	}
	for group, records := range recordByMedia {
		if group == "" {
			result.Records = append(result.Records, records...)
			continue
		}
		rootRecord := records[0]
		if len(records) > 1 {
			for _, record := range records[1:] {
				rootRecord.Links = append(rootRecord.Links, record.Links...)
				rootRecord.Files = append(rootRecord.Files, record.Files...)
				rootRecord.TextContent += "\n" + record.TextContent
			}
		}
		result.Records = append(result.Records, rootRecord)
	}
	ts.log.Infof("Resolving actual file urls")
	for _, record := range result.Records {
		for _, file := range record.Files {
			ts.log.Debugf("Loading file for %s", file)
			fileURL, err := ts.bot.GetFile(file.FileId)
			if err != nil {
				ts.log.Errorf("Cannot get actual file url for %s: %s", file.FileId, err)
				continue
			}
			file.FileUrl = ts.bot.GetUrl(fileURL)
		}
	}
	return result
}

func (ts *TelegramSource) fetchLoop() {
	for {
		updates := ts.waitForUpdate()
		ts.log.Infof("%d new updates.", len(updates))
		if len(updates) == 0 {
			continue
		}
		recordUpdates := []*telegram.Update{}
		requestUpdates := []*telegram.Update{}
		for _, upd := range updates {
			if isRecordUpdate(upd) {
				recordUpdates = append(recordUpdates, upd)
			} else {
				requestUpdates = append(requestUpdates, upd)
			}
		}
		if len(recordUpdates) > 0 {
			ts.log.Infof("%d record updates", len(recordUpdates))
			ts.records <- ts.groupRecords(recordUpdates)
		}
		if len(requestUpdates) > 0 {
			ts.log.Infof("%d request updates", len(requestUpdates))
			for _, request := range ts.groupRequests(requestUpdates) {
				ts.requests <- request
			}
		}
	}
}

func (ts *TelegramSource) waitForUpdate() []*telegram.Update {
	ts.log.Infof("Waiting for a new telegram update...")
	return collections.IterateSlice(
		optional.
			OfError(ts.bot.GetUpdates(int64(0), ts.cursor, 100, 100, []string{})).
			OrElse([]*telegram.Update{})).
		Filter(func(u *telegram.Update) bool {
			return u.Message != nil
		}).
		Peek(func(u *telegram.Update) {
			if ts.cursor <= u.UpdateID {
				ts.cursor = u.UpdateID + 1
			}
		}).Collect()
}

func (ts *TelegramSource) GetRecords() <-chan *rpb.RecordSet {
	return ts.records
}

func (ts *TelegramSource) GetRequest() <-chan *rpb.Request {
	return ts.requests
}

func isRecordUpdate(upd *telegram.Update) bool {
	msg := upd.Message
	return msg == nil || msg.ForwardFromChat != nil || msg.MediaGroupID != "" ||
		msg.Video != nil || msg.Photo != nil || msg.Audio != nil || msg.Voice != nil
}

func updateToRecords(upd *telegram.Update) (*rpb.Record, []*rpb.UserMetadata) {
	msg := upd.Message
	users := []*rpb.UserMetadata{}
	result := &rpb.Record{
		Source: &rpb.Source{
			SenderId:  fmt.Sprintf("%d", msg.From.ID),
			ChannelId: fmt.Sprintf("%d", msg.Chat.ID),
			MessageId: fmt.Sprintf("%d", msg.MessageID),
			Type:      rpb.SourceType_TELEGRAM,
		},
		Time: msg.Date,
	}
	users = append(users, &rpb.UserMetadata{
		Id:       fmt.Sprintf("%d", msg.From.ID),
		Username: msg.From.Username,
	})

	if msg.ForwardFromMessageID != 0 {
		result.Parent = &rpb.Source{
			MessageId: fmt.Sprintf("%d", msg.ForwardFromMessageID),
		}
	}
	if msg.ForwardFromChat != nil {
		if result.Parent == nil {
			result.Parent = &rpb.Source{}
		}
		users = append(users, &rpb.UserMetadata{
			Id:       fmt.Sprintf("%d", msg.ForwardFromChat.ID),
			Username: msg.ForwardFromChat.Username,
			Quotes:   []string{msg.ForwardFromChat.Title},
		})
		result.Parent.ChannelId = fmt.Sprintf("%d", msg.ForwardFromChat.ID)
	}
	if result.Parent != nil {
		result.Parent.Type = rpb.SourceType_TELEGRAM
	}

	textContent := []string{}
	if len(msg.Caption) != 0 {
		textContent = append(textContent, msg.Caption)
	}
	if len(msg.Text) != 0 {
		textContent = append(textContent, msg.Text)
	}
	result.TextContent = strings.Replace(strings.Join(textContent, "\n"), "\n\n", "\n", -1)
	for _, e := range msg.Entities {
		if (e.Type == "url" || e.Type == "text_link") && e.URL != "" {
			result.Links = append(result.Links, e.URL)
		}
	}
	if msg.Video != nil {
		result.Files = append(result.Files, &rpb.File{FileId: msg.Video.FileID})
	}
	if msg.Photo != nil {
		result.Files = append(result.Files, &rpb.File{
			FileId: telegram.GetLargestImage(msg.Photo).FileID,
		})
	}
	if msg.Audio != nil {
		result.Files = append(result.Files, &rpb.File{FileId: msg.Audio.FileID})
	}
	if msg.Voice != nil {
		result.Files = append(result.Files, &rpb.File{FileId: msg.Voice.FileID})
	}
	result.Links = collections.Unique(append(result.Links, util.FindWebLinks(result.TextContent)...))
	sort.Strings(result.Links)

	return result, users
}
