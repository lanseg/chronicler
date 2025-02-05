package twitter

import (
	"chronicler/adapter"
	"chronicler/common"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

var (
	tweetFields = []string{
		"article", "attachments", "author_id", "card_uri", "community_id", "context_annotations",
		"conversation_id", "created_at", "display_text_range", "edit_controls",
		"edit_history_tweet_ids", "entities", "geo", "id", "in_reply_to_user_id", "lang",
		"media_metadata",
		// "non_public_metrics", "note_tweet", "organic_metrics", "promoted_metrics", "public_metrics",
		"possibly_sensitive", "referenced_tweets",
		"reply_settings", "scopes", "source", "text", "withheld",
	}
	tweetExpansions = []string{
		"article.cover_media", "article.media_entities", "attachments.media_keys",
		"attachments.media_source_tweet", "attachments.poll_ids", "author_id",
		"edit_history_tweet_ids", "entities.mentions.username", "geo.place_id",
		"in_reply_to_user_id", "entities.note.mentions.username", "referenced_tweets.id",
		"referenced_tweets.id.author_id",
	}
	tweetMediaFields = []string{
		"url", "height", "width", "media_key", "variants",
	}
	tweetUserFields = []string{
		"id", "name", "username", "created_at", "description", "entities", "location", "url",
	}
)

type Client interface {
	GetTweets(ids []string) (*Response[Tweet], error)
	GetConversation(conversationId string) ([]*Response[Tweet], error)
}

type ClientImpl struct {
	Client

	httpClient adapter.HttpClient
	token      string

	logger *common.Logger
}

func NewClient(httpClient adapter.HttpClient, token string) Client {
	return &ClientImpl{
		token:      token,
		httpClient: httpClient,
		logger:     common.NewLogger("twitter"),
	}
}

func (c *ClientImpl) newRequest(url *url.URL) *http.Request {
	return &http.Request{
		Method: "GET",
		URL:    url,
		Header: http.Header{
			"Authorization": {fmt.Sprintf("Bearer %s", c.token)},
			"Content-Type":  {"application/json"},
		},
	}
}

func unmarshalResponse[T any](bytes []byte) (*Response[T], error) {
	result := &Response[T]{
		Includes: &Includes{},
		Meta:     &Metadata{},
	}
	err := json.Unmarshal(bytes, result)
	return result, err
}

func (c *ClientImpl) performRequest(url *url.URL) ([]byte, error) {
	resp, err := c.httpClient.Do(c.newRequest(url))
	if err != nil {
		return nil, err
	}
	return io.ReadAll(resp.Body)
}

func (c *ClientImpl) getConversationPage(conversationId string, paginationToken string) (*Response[Tweet], error) {
	url := &url.URL{
		Scheme: "https",
		Host:   "api.twitter.com",
		Path:   "/2/tweets/search/recent",
		RawQuery: fmt.Sprintf("query=conversation_id:%s&tweet.fields=%s&expansions=%s&user.fields=%s&media.fields=%s&max_results=100",
			url.QueryEscape(conversationId),
			url.QueryEscape(strings.Join(tweetFields, ",")),
			url.QueryEscape(strings.Join(tweetExpansions, ",")),
			url.QueryEscape(strings.Join(tweetUserFields, ",")),
			url.QueryEscape(strings.Join(tweetMediaFields, ","))),
	}
	if paginationToken != "" {
		url.RawQuery = fmt.Sprintf("%s&pagination_token=%s", url.RawQuery, paginationToken)
	}
	data, err := c.performRequest(url)
	if err != nil {
		return nil, err
	}
	return unmarshalResponse[Tweet](data)
}

func (c *ClientImpl) GetTweets(ids []string) (*Response[Tweet], error) {
	if len(ids) == 0 {
		return &Response[Tweet]{
			Data:   []*Tweet{},
			Errors: []*Error{},
			Includes: &Includes{
				Media:  []*Media{},
				Tweets: []*Tweet{},
				Users:  []*User{},
			},
			Meta: &Metadata{ResultCount: 0},
		}, nil
	}
	data, err := c.performRequest(&url.URL{
		Scheme: "https",
		Host:   "api.twitter.com",
		Path:   "/2/tweets",
		RawQuery: fmt.Sprintf("ids=%s&tweet.fields=%s&expansions=%s&media.fields=%s",
			url.QueryEscape(strings.Join(ids, ",")),
			url.QueryEscape(strings.Join(tweetFields, ",")),
			url.QueryEscape(strings.Join(tweetExpansions, ",")),
			url.QueryEscape(strings.Join(tweetMediaFields, ","))),
	})
	if err != nil {
		return nil, err
	}
	return unmarshalResponse[Tweet](data)
}

func (c *ClientImpl) GetConversation(conversationId string) ([]*Response[Tweet], error) {
	pagingToken := ""
	responses := []*Response[Tweet]{}
	for {
		result, err := c.getConversationPage(conversationId, pagingToken)
		if err != nil {
			c.logger.Errorf("Cannot load tweets from conversation %s with paging token %s: %s",
				conversationId, pagingToken, err)
			if len(responses) == 0 {
				return nil, err
			} else {
				break
			}
		}
		if result.Meta.ResultCount == 0 {
			c.logger.Infof("Conversation with id %s not found", conversationId)
			return []*Response[Tweet]{}, nil
		}
		responses = append(responses, result)
		pagingToken = result.Meta.NextToken
		c.logger.Infof("Loaded pages: %d, next page: %s", len(responses), pagingToken)
		if len(result.Data) == 0 || pagingToken == "" {
			break
		}
	}
	c.logger.Infof("Loaded %d pages from conversation %s", len(responses), conversationId)
	return responses, nil
}
