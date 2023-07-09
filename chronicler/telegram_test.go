package chronicler

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"chronicler/telegram"
)

// type Bot interface {
// 	GetUpdates(chatID int64, offset int64, limit int64, timeout int64, allowedUpdates []string) ([]*Update, error)
// 	SendMessage(chatID int64, replyId int64, text string) (*Message, error)
// 	GetFile(fileID string) (*File, error)
// 	GetUrl(file *File) string
// }

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
	bytes, err := os.ReadFile(filepath.Join("testdata", datafile))
	if err != nil {
		return nil, err
	}

	updates := []*telegram.Update{}
	err = json.Unmarshal(bytes, &updates)
	if err != nil {
		return nil, err
	}

	return &FakeBot{
		response: updates,
	}, nil
}

func TestRequestResponse(t *testing.T) {
	t.Run("Parse updates", func(t *testing.T) {
		bot, err := NewFakeBot("telegram_one_update.json")
		if err != nil {
			t.Errorf("Cannot create new fake bot for file \"telegram_one_update.json\": %s", err)
		}
		tg := NewTelegramChronicler(bot)
		ups := <-tg.GetRecords()
		fmt.Println(ups)
	})
}
