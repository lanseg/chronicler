package pikabu

import (
	"bytes"
	"chronicler/common"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"golang.org/x/text/encoding/charmap"
)

const (
	commentBatchSize = 200
)

type HttpClient interface {
	Do(request *http.Request) (*http.Response, error)
}

type CommentData struct {
	Id   int    `json:"id"`
	Html string `json:"html"`
}

type CommentResponse struct {
	Result      bool           `json:"result"`
	Message     string         `json:"message"`
	MessageCode int            `json:"message_code"`
	Data        []*CommentData `json:"data"`
}

type Client struct {
	httpClient HttpClient
	logger     *common.Logger
}

func NewClient(httpClient HttpClient) *Client {
	return &Client{
		httpClient: httpClient,
		logger:     common.NewLogger("PikabuClient"),
	}
}

func (c *Client) getComments(ids []string, from int, to int) (*CommentResponse, error) {
	requestBody := fmt.Sprintf("action=get_comments_by_ids&ids=%s", strings.Join(ids[from:to], ","))
	requestUrl, err := url.Parse("https://pikabu.ru/ajax/comments_actions.php")
	if err != nil {
		return nil, err
	}
	request := &http.Request{
		Method: "POST",
		URL:    requestUrl,
		Body:   io.NopCloser(bytes.NewReader([]byte(requestBody))),
		Header: http.Header{
			"Content-type": []string{"application/x-www-form-urlencoded"},
		},
	}
	resp, err := c.httpClient.Do(request)
	if err != nil {
		return nil, err
	}
	result, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	cr := &CommentResponse{}
	if err = json.Unmarshal(result, cr); err != nil {
		return nil, err
	}
	return cr, nil
}

func (c *Client) GetComments(ids []string) ([]*CommentData, error) {
	result := []*CommentData{}
	for i := 0; i < len(ids); i += commentBatchSize {
		end := i + commentBatchSize
		if end >= len(ids) {
			end = len(ids)
		}
		c.logger.Debugf("Loading comments [%4d to %4d] of %4d", i, end, len(ids))
		batch, err := c.getComments(ids, i, end)
		if err != nil {
			return nil, err
		}
		result = append(result, batch.Data...)
	}
	return result, nil
}

func (c *Client) GetPost(id string) (string, error) {
	requestUrl, err := url.Parse(fmt.Sprintf("https://pikabu.ru/story/_%s", id))
	if err != nil {
		return "", nil
	}
	request := &http.Request{
		Method: "GET",
		URL:    requestUrl,
	}
	resp, err := c.httpClient.Do(request)
	if err != nil {
		return "", err
	}
	data, err := io.ReadAll(charmap.Windows1251.NewDecoder().Reader(resp.Body))
	if err != nil {
		return "", err
	}
	return string(data), nil
}
