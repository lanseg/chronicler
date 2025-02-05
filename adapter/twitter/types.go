package twitter

import (
	"fmt"
	"strings"
)

type Error struct {
	Title  string `json:"title"`
	Type   string `json:"type"`
	Detail string `json:"detail"`
	Status int32  `json:"status"`
}

type Includes struct {
	Media  []*Media `json:"media"`
	Tweets []*Tweet `json:"tweets"`
	Users  []*User  `json:"users"`
}

type Attachment struct {
	MediaKeys []string `json:"media_keys"`
}

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
	return fmt.Sprintf("Media {key: %s, url: %s, size: %dx%d, variants: %v}",
		m.MediaKey, m.Url, m.Width, m.Height, m.Variants)
}

type User struct {
	Id       string `json:"id"`
	Name     string `json:"name"`
	Username string `json:"username"`
}

type ReferencedTweet struct {
	Id   string `json:"id"`
	Type string `json:"type"`
}

type Url struct {
	ExpandedUrl string `json:"expanded_url"`
	MediaKey    string `json:"media_key"`
}

type Entity struct {
	Urls []Url `json:"urls"`
}

type Tweet struct {
	Id             string            `json:"id"`
	Text           string            `json:"text"`
	Created        string            `json:"created_at"`
	ConversationId string            `json:"conversation_id"`
	AuthorId       string            `json:"author_id"`
	Attachments    Attachment        `json:"attachments"`
	Reference      []ReferencedTweet `json:"referenced_tweets"`
	Entities       Entity            `json:"entities"`
}

func (t Tweet) String() string {
	return fmt.Sprintf(
		"Tweet {id: %s, text: %s, created: %s, author: %s, attachments:%s, references:%s}",
		t.Id, t.Text, t.Created, t.AuthorId, t.Attachments, t.Reference)
}

type Metadata struct {
	NextToken   string `json:"next_token"`
	ResultCount uint64 `json:"result_count"`
}

type Response[T any] struct {
	Data     []*T      `json:"data"`
	Errors   []*Error  `json:"errors"`
	Includes *Includes `json:"includes"`
	Meta     *Metadata `json:"meta"`
}

func (r Response[T]) String() string {
	result := make([]string, len(r.Data))
	for i, tweet := range r.Data {
		result[i] = fmt.Sprintf("%v", tweet)
	}
	return fmt.Sprintf("Response {data: [%s]}", strings.Join(result, ", "))
}
