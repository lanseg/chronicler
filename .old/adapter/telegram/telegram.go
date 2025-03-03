package telegram

import (
	"fmt"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"

	col "github.com/lanseg/golang-commons/collections"
	cm "github.com/lanseg/golang-commons/common"
	"github.com/lanseg/tgbot"

	"chronicler/adapter"
	rpb "chronicler/records/proto"
	"chronicler/util"
)

type telegramAdapter struct {
	adapter.Adapter

	logger        *cm.Logger
	bot           tgbot.TelegramBot
	storagePrefix *url.URL
	api           *tgbot.TelegramApi
	cursor        int64
}

func NewTelegramAdapter(bot tgbot.TelegramBot, storagePrefix *url.URL) adapter.Adapter {
	return &telegramAdapter{
		logger:        cm.NewLogger("Telegramadapter.Adapter"),
		bot:           bot,
		api:           tgbot.NewTelegramApi(bot),
		storagePrefix: storagePrefix,
		cursor:        0,
	}
}

func (ts *telegramAdapter) resolveFileUrls(rs *rpb.RecordSet) {
	for _, record := range rs.Records {
		for _, file := range record.Files {
			ts.logger.Debugf("Resolving file url for %s", file)
			fileURL, err := ts.api.GetFile(&tgbot.GetFileRequest{
				FileID: file.FileId,
			})
			if err != nil {
				ts.logger.Errorf("Cannot get actual file url for %s: %s", file.FileId, err)
				continue
			}

			ts.logger.Debugf("Resolved file path to %s", fileURL.Result.FilePath)
			actualUrl := ts.storagePrefix.JoinPath(fileURL.Result.FilePath)
			ts.logger.Infof("Resolved file url for %s: %s", file.FileId, actualUrl.String())
			file.FileUrl = actualUrl.String()
		}
	}
}

func (ts *telegramAdapter) getUpdates() []*tgbot.Update {
	result, err := ts.api.GetUpdates(
		&tgbot.GetUpdatesRequest{
			Limit:          100,
			Offset:         ts.cursor,
			Timeout:        100,
			AllowedUpdates: []string{},
		})
	if err != nil {
		ts.logger.Errorf("Error: %s", err)
		return []*tgbot.Update{}
	}
	return result.Result
}

func (ts *telegramAdapter) waitForUpdate() []*tgbot.Update {
	return col.SliceStream(ts.getUpdates()).
		Peek(func(u *tgbot.Update) {
			if ts.cursor <= u.UpdateID {
				ts.cursor = u.UpdateID + 1
			}
		}).
		Filter(func(u *tgbot.Update) bool {
			return u.Message != nil
		}).Collect()
}

func (ts *telegramAdapter) FindSources(_ *rpb.Record) []*rpb.Source {
	return []*rpb.Source{}
}

func (ts *telegramAdapter) GetResponse(request *rpb.Request) []*rpb.Response {
	updates := ts.waitForUpdate()
	if len(updates) == 0 {
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
	channel := message.Target.ChannelId
	msgid, _ := strconv.Atoi(message.Target.MessageId)
	content := string(message.Content)
	_, err := ts.api.SendMessage(&tgbot.SendMessageRequest{
		ChatID: channel,
		ReplyParameters: &tgbot.ReplyParameters{
			MessageID: int64(msgid),
			ChatID:    channel,
		},
		Text: content,
	})
	if err == nil {
		ts.logger.Infof("Responded to channel(%s)/user(%d): %s", channel, msgid, content)
	} else {
		ts.logger.Infof("Failed to respond to channel(%s)/user(%d): %s", channel, msgid, err)
	}
}

// ----------

func groupRecords(updates []*tgbot.Update) []*rpb.RecordSet {
	grouped := col.GroupBy(updates, func(u *tgbot.Update) string {
		if u.Message == nil {
			return ""
		}
		return u.Message.MediaGroupID
	})
	result := []*rpb.RecordSet{}
	for _, v := range grouped {
		record, users := updateToRecords(v)
		rs := &rpb.RecordSet{
			Id:           cm.UUID4For(record.Source),
			Records:      []*rpb.Record{record},
			UserMetadata: users,
		}
		result = append(result, rs)
	}
	return result
}

func chatToMetadata(c *tgbot.Chat) *rpb.UserMetadata {
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

func userToMetadata(u *tgbot.User) *rpb.UserMetadata {
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

func videoToFile(v *tgbot.Video) *rpb.File {
	return &rpb.File{
		FileId: v.FileID,
	}
}

func photoToFile(photos []*tgbot.PhotoSize) *rpb.File {
	largestPhoto := getLargestImage(photos)
	return &rpb.File{
		FileId: largestPhoto.FileID,
	}
}

func audioToFile(audio *tgbot.Audio) *rpb.File {
	return &rpb.File{
		FileId: audio.FileID,
	}
}

func voiceToFile(voice *tgbot.Voice) *rpb.File {
	return &rpb.File{
		FileId: voice.FileID,
	}
}

func documentToFile(doc *tgbot.Document) *rpb.File {
	return &rpb.File{
		FileId: doc.FileID,
	}
}

func videoNoteToFile(note *tgbot.VideoNote) *rpb.File {
	return &rpb.File{
		FileId: note.FileID,
	}
}

func entitiesToLinks(ets []*tgbot.MessageEntity) []string {
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

func updateToRecords(upds []*tgbot.Update) (*rpb.Record, []*rpb.UserMetadata) {
	result := &rpb.Record{
		FetchTime: time.Now().Unix(),
	}
	users := map[string]*rpb.UserMetadata{}

	for _, vv := range upds {
		msg := vv.Message
		users[fmt.Sprintf("%d", msg.Chat.ID)] = chatToMetadata(msg.Chat)
		result.Source = toSource(msg.MessageID, msg.Chat.ID, msg.From.ID)

		if result.Time == 0 || result.Time > msg.Date {
			result.Time = msg.Date
		}

		if fwd := msg.ForwardOrigin; fwd != nil {
			msgId := int64(0)
			chatId := int64(0)
			userId := int64(0)
			if fwdUser := fwd.SenderUser; fwdUser != nil {
				users[fmt.Sprintf("%d", fwdUser.ID)] = userToMetadata(fwdUser)
				userId = fwdUser.ID
			}
			if fwdChat := fwd.SenderChat; fwdChat != nil {
				users[fmt.Sprintf("%d", fwdChat.ID)] = chatToMetadata(fwdChat)
				chatId = fwdChat.ID
			}
			if fwdChannel := fwd.Chat; fwdChannel != nil {
				users[fmt.Sprintf("%d", fwdChannel.ID)] = chatToMetadata(fwdChannel)
				chatId = fwdChannel.ID
				msgId = fwd.MessageID
			}
			result.Parent = toSource(msgId, chatId, userId)
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

		if msg.Document != nil {
			result.Files = append(result.Files, documentToFile(msg.Document))
		}

		if msg.VideoNote != nil {
			result.Files = append(result.Files, videoNoteToFile(msg.VideoNote))
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
	result.Links = col.Unique(
		append(result.Links, util.FindWebLinks(result.TextContent)...))

	sort.Strings(result.Links)

	userData := col.Values(users)
	sort.Slice(userData, func(i int, j int) bool {
		return userData[j].Id < userData[i].Id
	})
	return result, userData
}

// UTILS
func getLargestImage(sizes []*tgbot.PhotoSize) *tgbot.PhotoSize {
	if len(sizes) == 0 {
		return nil
	}

	var result *tgbot.PhotoSize = sizes[0]
	resultSize := int64(0)
	for _, photo := range sizes {
		size := photo.Width * photo.Height
		if size > resultSize {
			result = photo
			resultSize = size
		}
	}
	return result
}
