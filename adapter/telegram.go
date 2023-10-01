package adapter

import (
	"fmt"
	"sort"
	"strconv"
	"strings"

	"chronicler/records"
	"chronicler/telegram"
    "chronicler/util"

	"github.com/lanseg/golang-commons/collections"
	"github.com/lanseg/golang-commons/optional"
    cm "github.com/lanseg/golang-commons/common"

	rpb "chronicler/records/proto"
)

type telegramAdapter struct {
	Adapter

	logger *cm.Logger
	bot    telegram.Bot
	cursor int64
}

func NewTelegramAdapter(bot telegram.Bot) Adapter {
	return &telegramAdapter{
		logger: cm.NewLogger("TelegramAdapter"),
		bot:    bot,
		cursor: 0,
	}
}

func (ts *telegramAdapter) resolveFileUrls(rs *rpb.RecordSet) {
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

func (ts *telegramAdapter) waitForUpdate() []*telegram.Update {
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

func (ts *telegramAdapter) GetResponse(request *rpb.Request) []*rpb.Response {
	updates := ts.waitForUpdate()
	if len(updates) == 0 {
		ts.logger.Debugf("No updates for request: %s", request)
		return []*rpb.Response{}
	}
	records := groupRecords(updates)
	ts.logger.Infof("%d new updates grouped into %d record sets.", len(updates), len(records))

	result := []*rpb.Response{}
	for _, rs := range records {
		ts.resolveFileUrls(rs)
		result = append(result, &rpb.Response{
			Request: &rpb.Request{
				Origin: rs.Records[0].Source,
			},
			Result: []*rpb.RecordSet{rs},
		})
	}
	return result
}

func (ts *telegramAdapter) SendMessage(message *rpb.Message) {
	channel, _ := strconv.Atoi(message.Target.ChannelId)
	msgid, _ := strconv.Atoi(message.Target.MessageId)
	content := string(message.Content)
	_, err := ts.bot.SendMessage(int64(channel), int64(msgid), content)
	if err == nil {
		ts.logger.Infof("Responded to channel(%d)/user(%d): %s", channel, msgid, content)
	} else {
		ts.logger.Infof("Failed to respond to channel(%d)/user(%d): %s", channel, msgid, err)
	}
}

// ----------

func groupRecords(updates []*telegram.Update) []*rpb.RecordSet {
	grouped := collections.GroupBy(updates, func(u *telegram.Update) string {
		if u.Message == nil {
			return ""
		}
		return u.Message.MediaGroupID
	})
	result := []*rpb.RecordSet{}
	for _, v := range grouped {
		record, users := updateToRecords(v)
		rs := records.NewRecordSet([]*rpb.Record{record}, users)
		rs.Id = cm.UUID4()
		result = append(result, rs)
	}
	return result
}

func chatToMetadata(c *telegram.Chat) *rpb.UserMetadata {
	quotes := []string{}
	if c.Title != "" {
		quotes = append(quotes, c.Title)
	}
	return &rpb.UserMetadata{
		Id:       fmt.Sprintf("%d", c.ID),
		Username: c.Username,
		Quotes:   quotes,
	}
}

func userToMetadata(u *telegram.User) *rpb.UserMetadata {
	quotes := []string{}
	if u.FirstName != "" || u.LastName != "" {
		quotes = append(quotes, strings.TrimSpace(fmt.Sprintf("%s %s", u.FirstName, u.LastName)))
	}
	return &rpb.UserMetadata{
		Id:       fmt.Sprintf("%d", u.ID),
		Username: u.Username,
		Quotes:   quotes,
	}
}

func videoToFile(v *telegram.Video) *rpb.File {
	return &rpb.File{
		FileId: v.FileID,
	}
}

func photoToFile(photos []*telegram.PhotoSize) *rpb.File {
	largestPhoto := telegram.GetLargestImage(photos)
	return &rpb.File{
		FileId: largestPhoto.FileID,
	}
}

func audioToFile(audio *telegram.Audio) *rpb.File {
	return &rpb.File{
		FileId: audio.FileID,
	}
}

func voiceToFile(voice *telegram.Voice) *rpb.File {
	return &rpb.File{
		FileId: voice.FileID,
	}
}

func entitiesToLinks(ets []*telegram.MessageEntity) []string {
	result := []string{}
	for _, e := range ets {
		if (e.Type == "text_link" || e.Type == "url") && e.URL != "" {
			result = append(result, e.URL)
		}
	}
	return result
}

func toSource(msg int64, chat int64, user int64) *rpb.Source {
	src := &rpb.Source{
		Type: rpb.SourceType_TELEGRAM,
	}
	if msg != 0 {
		src.MessageId = fmt.Sprintf("%d", msg)
	}
	if chat != 0 {
		src.ChannelId = fmt.Sprintf("%d", chat)
	}
	if user != 0 {
		src.SenderId = fmt.Sprintf("%d", user)
	}
	return src
}

func updateToRecords(upds []*telegram.Update) (*rpb.Record, []*rpb.UserMetadata) {
	result := &rpb.Record{}
	users := map[string]*rpb.UserMetadata{}

	for _, vv := range upds {
		msg := vv.Message
		users[fmt.Sprintf("%d", msg.Chat.ID)] = chatToMetadata(msg.Chat)
		result.Source = toSource(msg.MessageID, msg.Chat.ID, msg.From.ID)

		if result.Time == 0 || result.Time > msg.Date {
			result.Time = msg.Date
		}

		if msg.ForwardFromChat != nil {
			ffId := int64(0)
			ffChat := msg.ForwardFromChat
			if msg.ForwardFrom != nil {
				ffId = msg.ForwardFrom.ID
			}
			users[fmt.Sprintf("%d", ffChat)] = chatToMetadata(ffChat)
			result.Parent = toSource(msg.ForwardFromMessageID, msg.ForwardFromChat.ID, ffId)
		}

		if msg.Video != nil {
			result.Files = append(result.Files, videoToFile(msg.Video))
		}

		if msg.Photo != nil {
			result.Files = append(result.Files, photoToFile(msg.Photo))
		}

		if msg.Audio != nil {
			result.Files = append(result.Files, audioToFile(msg.Audio))
		}

		if msg.Voice != nil {
			result.Files = append(result.Files, voiceToFile(msg.Voice))
		}

		if msg.Entities != nil {
			result.Links = append(result.Links, entitiesToLinks(msg.Entities)...)
		}

		if msg.Caption != "" {
			result.TextContent += strings.TrimSpace(msg.Caption) + "\n"
		}

		if msg.Text != "" {
			result.TextContent += strings.TrimSpace(msg.Text) + "\n"
		}
	}

	result.TextContent = strings.TrimSpace(result.TextContent)
	result.Links = collections.Unique(
		append(result.Links, util.FindWebLinks(result.TextContent)...))

	sort.Strings(result.Links)

	userData := collections.Values(users)
	sort.Slice(userData, func(i int, j int) bool {
		return userData[j].Id < userData[i].Id
	})
	return result, userData
}
