package twitter

import (
	"fmt"
	"strings"
)

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
	return fmt.Sprintf("Media {key: %s, url: %s, size: %dx%d, variants: %s}", m.MediaKey, m.Url, m.Width, m.Height, m.Variants)
}

type Includes struct {
	Media []Media `json:"media"`
}

type Attachment struct {
	MediaKeys []string `json:"media_keys"`
}

type ReferencedTweet struct {
	Id   string `json: "id"`
	Type string `json: "type"`
}

type Tweet struct {
	Id          string            `json:"id"`
	Text        string            `json:"text"`
	Created     string            `json:"created_at"`
	Author      string            `json:"author_id"`
	Attachments Attachment        `json:"attachments"`
	Reference   []ReferencedTweet `json:"referenced_tweets"`
	Media       []*TwitterMedia
}

func (t Tweet) String() string {
	return fmt.Sprintf("Tweet {id: %s, text: %s, created: %s, author: %s, attachments:%s, references:%s, media:%s}",
		t.Id, t.Text, t.Created, t.Author, t.Attachments, t.Reference, t.Media)
}

type Response struct {
	Data     []*Tweet  `json:"data"`
	Includes *Includes `json:"includes"`
}

func (r Response) String() string {
	result := make([]string, len(r.Data))
	for i, tweet := range r.Data {
		result[i] = tweet.String()
	}
	return fmt.Sprintf("Response {data: [%s]}", strings.Join(result, ", "))
}
