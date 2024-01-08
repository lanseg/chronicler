package telegram

import (
	"bytes"
	"fmt"
	"io"

	"encoding/json"
	"net/http"
	"net/url"

	cm "github.com/lanseg/golang-commons/common"
	tgbot "github.com/lanseg/tgbot"
)

type Response[T any] struct {
	Ok          bool   `json:"ok"`
	ErrorCode   int64  `json:"error_code"`
	Description string `json:"description"`

	Result T `json:"result"`
}

type Bot interface {
	GetUpdates(chatID int64, offset int64, limit int64, timeout int64, allowedUpdates []string) ([]*tgbot.Update, error)
	SendMessage(chatID int64, replyId int64, text string) (*tgbot.Message, error)
	GetFile(fileID string) (*tgbot.File, error)
	GetUrl(file *tgbot.File) string
}

type BotImpl struct {
	Bot

	httpClient *http.Client
	token      string

	logger *cm.Logger
}

func NewBot(token string) Bot {
	return &BotImpl{
		token:      token,
		httpClient: &http.Client{},
		logger:     cm.NewLogger("telegram"),
	}
}

func (b *BotImpl) queryApi(apiMethod string, params url.Values) ([]byte, error) {
	resp, err := http.PostForm(fmt.Sprintf("https://api.telegram.org/bot%s/%s", b.token, apiMethod), params)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var buf bytes.Buffer
	_, err = io.Copy(&buf, resp.Body)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func queryAndUnmarshal[T any](b *BotImpl, apiMethod string, params url.Values) (T, error) {
	var zero T
	resultBytes, err := b.queryApi(apiMethod, params)
	if err != nil {
		return zero, err
	}
	response := &Response[T]{}
	if err = json.Unmarshal(resultBytes, response); err != nil {
		return zero, fmt.Errorf("Cannot unmarshal the response: %s", err)
	}
	if !response.Ok {
		return zero, fmt.Errorf("Request \"%s\" completed with error %d: %s",
			apiMethod, response.ErrorCode, response.Description)
	}
	return response.Result, nil
}

func (b *BotImpl) GetUpdates(chatID int64, offset int64, limit int64, timeout int64, allowedUpdates []string) ([]*tgbot.Update, error) {
	params := url.Values{}
	params.Set("chat_id", fmt.Sprintf("%d", chatID))
	params.Set("offset", fmt.Sprintf("%d", offset))
	params.Set("limit", fmt.Sprintf("%d", limit))
	params.Set("timeout", fmt.Sprintf("%d", timeout))
	for _, upd := range allowedUpdates {
		params.Add("allowed_updates", upd)
	}

	result, err := queryAndUnmarshal[[]*tgbot.Update](b, "getUpdates", params)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (b *BotImpl) SendMessage(chatID int64, replyId int64, text string) (*tgbot.Message, error) {
	params := url.Values{}
	params.Set("chat_id", fmt.Sprintf("%d", chatID))
	params.Set("text", text)
	if replyId > 0 {
		params.Set("reply_to_message_id", fmt.Sprintf("%d", replyId))
	}

	result, err := queryAndUnmarshal[*tgbot.Message](b, "sendMessage", params)
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (b *BotImpl) GetFile(fileID string) (*tgbot.File, error) {
	params := url.Values{}
	params.Set("file_id", fileID)

	result, err := queryAndUnmarshal[*tgbot.File](b, "getFile", params)
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (b *BotImpl) GetUrl(file *tgbot.File) string {
	return fmt.Sprintf("https://api.telegram.org/file/bot%s/%s",
		b.token, file.FilePath)
}
