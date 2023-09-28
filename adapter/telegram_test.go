package adapter

import (
	"fmt"
	"testing"

	"chronicler/telegram"

	rpb "chronicler/records/proto"
)

const (
	testingUuid = "1a468cef-1368-408a-a20b-86b32d94a460"
)

type FakeBot struct {
	telegram.Bot

	responded bool
	response  []*telegram.Update
}

func (b *FakeBot) GetUpdates(chatID int64, offset int64, limit int64, timeout int64, allowedUpdates []string) ([]*telegram.Update, error) {
	if b.responded {
		return []*telegram.Update{}, nil
	}
	b.responded = true
	return b.response, nil
}

func (b *FakeBot) SendMessage(chatID int64, replyId int64, text string) (*telegram.Message, error) {
	return &telegram.Message{}, nil
}

func (b *FakeBot) GetFile(fileID string) (*telegram.File, error) {
	return &telegram.File{
		FileID: fileID,
	}, nil
}

func (b *FakeBot) GetUrl(file *telegram.File) string {
	return fmt.Sprintf("https://telegram/url/%s", file)
}

func NewFakeBot(datafile string) (telegram.Bot, error) {
	updates := []*telegram.Update{}
	if err := readJson(datafile, &updates); err != nil {
		return nil, err
	}
	return &FakeBot{response: updates}, nil
}

func TestRequestResponse(t *testing.T) {
	for _, tc := range []struct {
		desc         string
		responseFile string
		resultFile   string
	}{
		{
			desc:         "Single update response",
			responseFile: "telegram_one_update.json",
			resultFile:   "telegram_one_update_record.json",
		},
		{
			desc:         "Multiple update response",
			responseFile: "telegram_multiple_updates.json",
			resultFile:   "telegram_multiple_updates_record.json",
		},
	} {
		t.Run(tc.desc, func(t *testing.T) {
			bot, err := NewFakeBot(tc.responseFile)
			if err != nil {
				t.Errorf("Cannot create new fake bot for file \"%s\": %s", tc.responseFile, err)
			}
			tg := NewTelegramAdapter(bot)
			tg.SubmitRequest(&rpb.Request{Id: testingUuid})
			ups := tg.GetResponse().Result[0]
			ups.Id = testingUuid

			want := &rpb.RecordSet{}
			if err = readJson(tc.resultFile, want); err != nil {
				t.Errorf("Cannot load json with an expected result \"%s\": %s", tc.resultFile, err)
			}
			if fmt.Sprintf("%+v", want) != fmt.Sprintf("%+v", ups) {
				t.Errorf("Expected result to be:\n%+v\nBut got:\n%+v", want, ups)
			}
		})
	}
}
