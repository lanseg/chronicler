package adapter

import (
	"fmt"
	"sort"
	"strconv"
	"strings"

	"chronicler/telegram"
	"chronicler/util"

	"github.com/lanseg/golang-commons/collections"
	"github.com/lanseg/golang-commons/optional"

	rpb "chronicler/proto/records"
)

type telegramSinkSource struct {
	SinkSource

	logger *util.Logger
	bot    telegram.Bot
	cursor int64
}

func NewTelegramAdapter(bot telegram.Bot) Adapter {
	tss := &telegramSinkSource{
		logger: util.NewLogger("TelegramAdapter"),
		bot:    bot,
		cursor: 0,
	}
	return NewAdapter("TelegramAdapter", tss, tss, true)
}

func (ts *telegramSinkSource) resolveFileUrls(rs *rpb.RecordSet) {
	for _, record := range rs.Records {
		for _, file := range record.Files {
			ts.logger.Debugf("Resolving file url for %s", file)
			fileURL, err := ts.bot.GetFile(file.FileId)
			if err != nil {
				ts.logger.Errorf("Cannot get actual file url for %s: %s", file.FileId, err)
				continue
			}
			file.FileUrl = ts.bot.GetUrl(fileURL)
		}
	}
}

func (ts *telegramSinkSource) waitForUpdate() []*telegram.Update {
	ts.logger.Infof("Waiting for a new telegram update...")
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

func (ts *telegramSinkSource) GetRequestedRecords(*rpb.Request) []*rpb.RecordSet {
	updates := ts.waitForUpdate()
	records := groupRecords(updates)
	ts.logger.Infof("%d new updates grouped into %d records.", len(updates), len(records))
	for _, rs := range records {
		ts.resolveFileUrls(rs)
	}
	return records
}

func (ts *telegramSinkSource) SendResponse(response *rpb.Response) {
	channel, _ := strconv.Atoi(response.Source.ChannelId)
	msgid, _ := strconv.Atoi(response.Source.MessageId)
	msg, err := ts.bot.SendMessage(int64(channel), int64(msgid), response.Content)
	if err == nil {
		ts.logger.Infof("Responded to channel(%d)/user(%d): %s", channel, msgid, response.Content)
	} else {
		ts.logger.Infof("Failed to respond to channel(%d)/user(%d): %s", channel, msg, err)
	}
}

//  ----------

func getUpdateSource(upd *telegram.Update) *rpb.Source {
	msg := upd.Message
	if msg == nil {
		return nil
	}
	return &rpb.Source{
		SenderId:  fmt.Sprintf("%d", msg.From.ID),
		ChannelId: fmt.Sprintf("%d", msg.Chat.ID),
		MessageId: fmt.Sprintf("%d", msg.MessageID),
		Type:      rpb.SourceType_TELEGRAM,
	}
}

func updateToRecords(upd *telegram.Update) (*rpb.Record, []*rpb.UserMetadata) {
	msg := upd.Message
	users := []*rpb.UserMetadata{}
	result := &rpb.Record{
		Source: getUpdateSource(upd),
		Time:   msg.Date,
		//      RawContent:
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

func groupRecords(updates []*telegram.Update) []*rpb.RecordSet {
	recordByMedia := map[string][]*rpb.Record{}
	userById := map[string]*rpb.UserMetadata{}

	recordByMedia[""] = []*rpb.Record{}
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

	allRecords := recordByMedia[""]
	aggregatedRecords := []*rpb.Record{}
	for group, records := range recordByMedia {
		if group == "" {
			continue
		}
		rootRecord := records[0]
		for _, record := range records[1:] {
			rootRecord.Links = append(rootRecord.Links, record.Links...)
			rootRecord.Files = append(rootRecord.Files, record.Files...)
			rootRecord.TextContent += "\n" + record.TextContent
		}
		aggregatedRecords = append(aggregatedRecords, rootRecord)
	}

	result := []*rpb.RecordSet{}
	metadata := collections.Values(userById)
	sort.Slice(metadata, func(i int, j int) bool {
		return metadata[i].Id > metadata[j].Id
	})
	for _, record := range append(allRecords, aggregatedRecords...) {
		result = append(result, &rpb.RecordSet{
			Request:      &rpb.Request{Source: record.Source},
			Records:      []*rpb.Record{record},
			UserMetadata: metadata,
		})
	}

	return result
}
