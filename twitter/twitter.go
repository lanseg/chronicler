package twitter

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"

	"chronicler/util"

	"github.com/lanseg/golang-commons/collections"
	"github.com/lanseg/golang-commons/optional"
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
	}
	tweetMediaFields = []string{
		"url", "height", "width", "media_key", "variants",
	}
	tweetUserFields = []string{
		"id", "name", "username",
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
	Media  []Media `json:"media"`
	Tweets []Tweet `json:"tweets"`
	Users  []User  `json:"users"`
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
}

func (t Tweet) String() string {
	return fmt.Sprintf(
		"Tweet {id: %s, text: %s, created: %s, author: %s, attachments:%s, references:%s}",
		t.Id, t.Text, t.Created, t.Author, t.Attachments, t.Reference)
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

type Response[T any] struct {
	Data     []T       `json:"data"`
	Includes *Includes `json:"includes"`
	Errors   []*Error  `json:"errors"`
	Meta     *Metadata `json:"meta"`
}

func (r Response[T]) String() string {
	result := make([]string, len(r.Data))
	for i, tweet := range r.Data {
		result[i] = fmt.Sprintf("%s", tweet)
	}
	return fmt.Sprintf("Response {data: [%s]}", strings.Join(result, ", "))
}

// The client
type Client interface {
	GetTweets(ids []string) (*Response[Tweet], error)
	GetConversation(conversationId string) (*Response[Tweet], error)
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

func getMissingMedia(response *Response[Tweet]) map[string]([]string) {
	mediaById := map[string]*Media{}
	for _, r := range response.Includes.Media {
		mediaById[r.MediaKey] = &r
	}
	result := map[string]([]string){}
	for _, tweet := range append(response.Data, response.Includes.Tweets...) {
		for _, attachedMediaKey := range tweet.Attachments.MediaKeys {
			if mediaById[attachedMediaKey] == nil || mediaById[attachedMediaKey].Url == "" {
				result[tweet.Id] = append(result[tweet.Id], attachedMediaKey)
			}
		}
	}
	return result
}

func getMissingUsers(response *Response[Tweet]) []string {
	users := collections.NewSet([]string{})
	for _, tweet := range response.Data {
		users.Add(tweet.Author)
	}
	for _, tweet := range response.Includes.Tweets {
		users.Add(tweet.Author)
	}
	for _, user := range response.Includes.Users {
		users.Remove(user.Id)
	}
	return users.Values()
}

func GetBestQualityMedia(media Media) *TwitterMedia {
	result := &TwitterMedia{}

	bitrate := int64(0)
	url := media.Url
	for _, v := range media.Variants {
		if bitrate <= v.Bitrate && v.Url != "" {
			url = v.Url
			bitrate = v.Bitrate
		}
	}
	size := media.Width * media.Height
	if size > result.Width*result.Height {
		result.Width = media.Width
		result.Height = media.Height
		result.Url = url
		result.Bitrate = bitrate
		result.Id = media.MediaKey
	}

	return result
}

func (c *ClientImpl) newRequest(url string) optional.Optional[*http.Request] {
	return optional.Map(optional.OfErrorNullable(http.NewRequest("GET", url, nil)),
		func(request *http.Request) *http.Request {
			request.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.token))
			request.Header.Set("Content-Type", "application/json")
			return request
		})
}

func unmarshalResponse[T any](bytes optional.Optional[[]byte]) optional.Optional[*Response[T]] {
	return optional.MapErr(bytes, func(b []byte) (*Response[T], error) {
		result := &Response[T]{
			Includes: &Includes{},
			Meta:     &Metadata{},
		}
		err := json.Unmarshal(b, result)
		return result, err
	})
}

func (c *ClientImpl) performRequest(url url.URL) optional.Optional[[]byte] {
	c.logger.Debugf("Api request: %s", url.String())
	return optional.MapErr(
		optional.MapErr(c.newRequest(url.String()), func(req *http.Request) (*http.Response, error) {
			return c.httpClient.Do(req)
		}),
		func(resp *http.Response) ([]byte, error) {
			defer resp.Body.Close()
			result, err := ioutil.ReadAll(resp.Body)
			return result, err
		})
}

func (c *ClientImpl) getUserInfo(ids []string) optional.Optional[*Response[User]] {
	url := url.URL{
		Scheme: "https",
		Host:   "api.twitter.com",
		Path:   "2/users",
		RawQuery: fmt.Sprintf("ids=%s&user.fields=%s",
			url.QueryEscape(strings.Join(ids, ",")),
			url.QueryEscape(strings.Join(tweetUserFields, ",")),
		),
	}
	return unmarshalResponse[User](c.performRequest(url))
}

func (c *ClientImpl) getMediaForTweets(ids []string) optional.Optional[[]Media] {
	url := url.URL{
		Scheme: "https",
		Host:   "api.twitter.com",
		Path:   "2/tweets",
		RawQuery: fmt.Sprintf("ids=%s&expansions=%s&media.fields=%s",
			url.QueryEscape(strings.Join(ids, ",")),
			url.QueryEscape("attachments.media_keys"),
			url.QueryEscape(strings.Join(tweetMediaFields, ","))),
	}
	return optional.Map(unmarshalResponse[Tweet](c.performRequest(url)),
		func(response *Response[Tweet]) []Media {
			return response.Includes.Media
		})
}

func (c *ClientImpl) getConversationPage(conversationId string, paginationToken string) optional.Optional[*Response[Tweet]] {
	url := url.URL{
		Scheme: "https",
		Host:   "api.twitter.com",
		Path:   "2/tweets/search/recent",
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
	return unmarshalResponse[Tweet](c.performRequest(url))
}

func (c *ClientImpl) GetTweets(ids []string) (*Response[Tweet], error) {
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
	return unmarshalResponse[Tweet](c.performRequest(url)).Get()
}

func (c *ClientImpl) GetConversation(conversationId string) (*Response[Tweet], error) {
	token := ""
	responses := []*Response[Tweet]{}
	for {
		result, err := c.getConversationPage(conversationId, token).Get()
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

	result := &Response[Tweet]{
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

	missingMedia := getMissingMedia(result)
	c.logger.Infof("Downloading missing media from tweets %s, for keys: %s",
		strings.Join(collections.Keys(missingMedia), ", "), collections.Values(missingMedia))
	if moreMedia, err := c.getMediaForTweets(collections.Keys(missingMedia)).Get(); err == nil {
		result.Includes.Media = append(result.Includes.Media, moreMedia...)
	} else {
		c.logger.Warningf("Failed loading media for %s: %s", collections.Keys(missingMedia))
	}
	missingMedia = getMissingMedia(result)
	c.logger.Infof("Missing media after the download: from tweets %s, for keys: %s",
		collections.Keys(missingMedia), collections.Values(missingMedia))

	missingUsers := getMissingUsers(result)
	c.logger.Infof("Downloading missing information for users: %s", missingUsers)
	if users, err := c.getUserInfo(missingUsers).Get(); err == nil {
		result.Includes.Users = append(result.Includes.Users, users.Data...)
	} else {
		c.logger.Warningf("Failed loading user information for %s", missingUsers)
	}
	missingUsers = getMissingUsers(result)
	c.logger.Infof("Missing information for users after download: %s", missingUsers)

	return result, nil
}
