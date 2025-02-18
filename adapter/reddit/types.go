package reddit

import (
	"encoding/json"
	"strconv"
)

type Thing[T any] struct {
	Id   string `json:"id"`
	Name string `json:"name"`
	Kind string `json:"kind"`
	Data T      `json:"data"`
}

type Listing[T any] struct {
	Before   string `json:"before"`
	After    string `json:"after"`
	Modhash  string `json:"modhash"`
	Children []T    `json:"children"`
}

// Votable contains attributes related to upvote or downvote.
type Votable struct {
	// The number of upvotes. (includes own)
	Ups int `json:"ups"`

	// The number of downvotes. (includes own)
	Downs int `json:"downs"`

	// true if thing is liked by the user, false if thing is disliked, null if the user has not
	// voted or you are not logged in.
	Likes bool `json:"likes"`
}

// Created contains time when the entity was created.
type Created struct {
	// The time of creation in local epoch-second format. ex: 1331042771.0
	Created float64 `json:"created"`

	// the time of creation in UTC epoch-second format. Note that neither of these ever have a non-zero fraction.
	CreatedUtc float64 `json:"created_utc"`
}

type Dimension struct {
	Height uint32 `json:"height"`
	Width  uint32 `json:"width"`
}

type Replies Thing[Listing[Thing[Entity]]]

type RepliesWrapper struct {
	Replies *Replies `json:"replies"`
	Empty   bool     `json:"empty"`
}

func (cr *RepliesWrapper) UnmarshalJSON(data []byte) error {
	if len(data) == 2 {
		cr.Empty = true
		return nil
	}
	cr.Replies = &Replies{}
	return json.Unmarshal(data, cr.Replies)
}

type FalseOrFloat float64

func (cr *FalseOrFloat) UnmarshalJSON(data []byte) error {
	result := FalseOrFloat(0)
	if data[0] == 'f' || data[0] == 'F' {
		*cr = result
		return nil
	}
	asFloat, err := strconv.ParseFloat(string(data), 64)
	if err != nil {
		return err
	}
	*cr = FalseOrFloat(asFloat)
	return nil
}

type RedditVideo struct {
	Dimension

	BitrateKbps       uint32 `json:"bitrate_kbps"`
	FallbackUrl       string `json:"fallback_url"`
	ScrubberMediaUrl  string `json:"scrubber_media_url"`
	DashUrl           string `json:"dash_url"`
	Duration          uint32 `json:"duration"`
	HLSUrl            string `json:"hls_url"`
	IsGif             bool   `json:"is_gif"`
	TranscodingStatus string `json:"transcoding_status"`
}

type Image struct {
	Dimension

	Url string `json:"url"`
}

type RedditImage struct {
	Id          string   `json:"id"`
	Source      *Image   `json:"source"`
	Resolutions []*Image `json:"resolutions"`
}

type Preview struct {
	Images []*RedditImage `json:"images"`
}

type Media struct {
	RedditVideo *RedditVideo `json:"reddit_video"`
}

type MediaVariant struct {
	Width  int    `json:"x"`
	Height int    `json:"y"`
	Url    string `json:"u"`
	Gif    string `json:"gif"`
	Mp4    string `json:"mp4"`
}

type MediaMetadata struct {
	MimeType string          `json:"m"`
	Status   string          `json:"status"`
	Previews []*MediaVariant `json:"p"`
	Source   *MediaVariant   `json:"s"`
}

// Entity is Reddit's comment or pos
type Entity struct {
	Votable
	Created

	Id             string `json:"id"`
	ParentId       string `json:"parent_id"`
	Name           string `json:"name"`
	Title          string `json:"title"`
	Body           string `json:"body"`
	BodyHtml       string `json:"body_html"`
	Author         string `json:"author"`
	AuthorFullName string `json:"author_fullname"`
	Subreddit      string `json:"subreddit"`
	Permalink      string `json:"permalink"`

	Replies       *RepliesWrapper           `json:"replies"`
	SecureMedia   *Media                    `json:"secure_media"`
	Media         *Media                    `json:"media"`
	MediaMetadata map[string]*MediaMetadata `json:"media_metadata"`
	Preview       *Preview                  `json:"preview"`
	Children      []string                  `json:"children"`
}
