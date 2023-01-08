package twitter

import (
	"fmt"
	"strings"

	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/url"

	"chronist/util"
)

var (
	tweetFields = []string{
		"attachments",
		"created_at",
		"author_id",
		"referenced_tweets",
		"conversation_id",
	}
	tweetExpansions = []string{
		"author_id",
		"attachments.media_keys",
		"referenced_tweets.id",
		"entities.mentions.username",
	}
	tweetMediaFields = []string{
		"url", "height", "width", "media_key", "variants",
	}
)

// Media types
type TwitterMedia struct {
	Width   int
	Height  int
	Bitrate int64
	Url     string
	Id      string
}

type MediaVariant struct {
	Bitrate     int64  `json:"bit_rate"`
	ContentType string `json:"content_type"`
	Url         string `json:"url"`
}

type Media struct {
	MediaKey string          `json:"media_key"`
	Url      string          `json:"url"`
	Width    int             `json:"width"`
	Height   int             `json:"height"`
	Variants []*MediaVariant `json:"variants"`
}

func (m Media) String() string {
	return fmt.Sprintf("Media {key: %s, url: %s, size: %dx%d, variants: %s}",
		m.MediaKey, m.Url, m.Width, m.Height, m.Variants)
}

type User struct {
	Id       string `json:"id"`
	Name     string `json:"name"`
	Username string `json:"username"`
}

type Includes struct {
	Media  []Media  `json:"media"`
	Tweets []*Tweet `json:"tweets"`
	Users  []User   `json:"users"`
}

type Attachment struct {
	MediaKeys []string `json:"media_keys"`
}

// Tweet types
type ReferencedTweet struct {
	Id   string `json: "id"`
	Type string `json: "type"`
}

type Tweet struct {
	Id             string            `json:"id"`
	Text           string            `json:"text"`
	Created        string            `json:"created_at"`
	ConversationId string            `json:"conversation_id"`
	Author         string            `json:"author_id"`
	Attachments    Attachment        `json:"attachments"`
	Reference      []ReferencedTweet `json:"referenced_tweets"`
	Media          []*TwitterMedia
}

func (t Tweet) String() string {
	return fmt.Sprintf(
		"Tweet {id: %s, text: %s, created: %s, author: %s, attachments:%s, references:%s, media:%s}",
		t.Id, t.Text, t.Created, t.Author, t.Attachments, t.Reference, t.Media)
}

type Error struct {
	ResourceId   string `json:"resource_id"`
	Parameter    string `json:"parameter"`
	ResourceType string `json:"resource_type"`
	Section      string `json:"section"`
	Title        string `json:"title"`
	Value        string `json:"value"`
	Detail       string `json:"detail"`
	Type         string `json:"type"`
}

type Metadata struct {
	NextToken   string `json:"next_token"`
	ResultCount uint64 `json:"result_count"`
}

type Response struct {
	Data     []*Tweet  `json:"data"`
	Includes *Includes `json:"includes"`
	Errors   []*Error  `json:"errors"`
	Meta     *Metadata `json:"meta"`
}

func (r Response) String() string {
	result := make([]string, len(r.Data))
	for i, tweet := range r.Data {
		result[i] = tweet.String()
	}
	return fmt.Sprintf("Response {data: [%s]}", strings.Join(result, ", "))
}

// The client
type Client interface {
	GetTweets(ids []string) (*Response, error)
	GetConversation(conversationId string) (*Response, error)
}

type ClientImpl struct {
	Client

	httpClient *http.Client
	token      string

	logger *util.Logger
}

func NewClient(token string) Client {
	return &ClientImpl{
		token:      token,
		httpClient: &http.Client{},
		logger:     util.NewLogger("twitter"),
	}
}

func getBestQualityMedia(medias []Media) *TwitterMedia {
	if len(medias) == 0 {
		return nil
	}
	result := &TwitterMedia{}

	for _, m := range medias {
		bitrate := int64(0)
		url := m.Url
		for _, v := range m.Variants {
			if bitrate <= v.Bitrate && v.Url != "" {
				url = v.Url
				bitrate = v.Bitrate
			}
		}
		size := m.Width * m.Height
		if size > result.Width*result.Height {
			result.Width = m.Width
			result.Height = m.Height
			result.Url = url
			result.Bitrate = bitrate
			result.Id = m.MediaKey
		}
	}

	return result
}

func (c *ClientImpl) newRequest(url string) (*http.Request, error) {
	request, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	request.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.token))
	request.Header.Set("Content-Type", "application/json")
	return request, nil
}

func (c *ClientImpl) performRequest(url url.URL) (*Response, error) {
	request, err := c.newRequest(url.String())
	if err != nil {
		return nil, err
	}

	response, err := c.httpClient.Do(request)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	bytes, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}
	result := &Response{
		Includes: &Includes{},
		Meta:     &Metadata{},
	}
	if err = json.Unmarshal(bytes, result); err != nil {
		return nil, err
	}
	mediaByKey := util.GroupBy(result.Includes.Media, func(m Media) string {
		return m.MediaKey
	})

	for _, tweet := range result.Data {
		for _, mk := range tweet.Attachments.MediaKeys {
			if medias, ok := mediaByKey[mk]; ok {
				tweet.Media = append(tweet.Media, getBestQualityMedia(medias))
			}
		}
	}
	return result, nil
}

func (c *ClientImpl) getConversationPage(conversationId string, paginationToken string) (*Response, error) {
	url := url.URL{
		Scheme: "https",
		Host:   "api.twitter.com",
		Path:   "2/tweets/search/recent",
		RawQuery: fmt.Sprintf("query=conversation_id:%s&tweet.fields=%s&expansions=%s&media.fields=%s&max_results=100",
			url.QueryEscape(conversationId),
			url.QueryEscape(strings.Join(tweetFields, ",")),
			url.QueryEscape(strings.Join(tweetExpansions, ",")),
			url.QueryEscape(strings.Join(tweetMediaFields, ","))),
	}
	if paginationToken != "" {
		url.RawQuery = fmt.Sprintf("%s&pagination_token=%s", url.RawQuery, paginationToken)
	}
	return c.performRequest(url)
}

func (c *ClientImpl) GetConversation(conversationId string) (*Response, error) {
	token := ""
	responses := []*Response{}
	for {
		result, err := c.getConversationPage(conversationId, token)
		if err != nil {
			c.logger.Errorf("Cannot load tweets from converation %s with token %s: %s",
				conversationId, token, err)
			if len(responses) == 0 {
				return nil, err
			} else {
				break
			}
		}
		responses = append(responses, result)
		token = result.Meta.NextToken
		c.logger.Infof("Loaded pages: %d, next page: %s", len(responses), token)
		if len(result.Data) == 0 || token == "" {
			break
		}
	}
	c.logger.Infof("Loaded %d pages from conversation %s", len(responses), conversationId)

	result := &Response{
		Includes: &Includes{},
		Meta:     &Metadata{},
	}
	for _, r := range responses {
		result.Data = append(result.Data, r.Data...)
		result.Errors = append(result.Errors, r.Errors...)
		result.Includes.Users = append(result.Includes.Users, r.Includes.Users...)
		result.Includes.Media = append(result.Includes.Media, r.Includes.Media...)
		result.Includes.Tweets = append(result.Includes.Tweets, r.Includes.Tweets...)
	}
	result.Meta.ResultCount = uint64(len(result.Data))
	return result, nil
}

func (c *ClientImpl) GetTweets(ids []string) (*Response, error) {
	url := url.URL{
		Scheme: "https",
		Host:   "api.twitter.com",
		Path:   "2/tweets",
		RawQuery: fmt.Sprintf("ids=%s&tweet.fields=%s&expansions=%s&media.fields=%s",
			url.QueryEscape(strings.Join(ids, ",")),
			url.QueryEscape(strings.Join(tweetFields, ",")),
			url.QueryEscape(strings.Join(tweetExpansions, ",")),
			url.QueryEscape(strings.Join(tweetMediaFields, ","))),
	}
	return c.performRequest(url)
}
