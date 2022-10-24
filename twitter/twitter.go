package twitter

import (
	"chronist/util"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
)

var (
	tweetFields = []string{
		"attachments",
		"created_at",
		"author_id",
		"referenced_tweets",
	}
	tweetExpansions = []string{
		"attachments.media_keys",
	}
	tweetMediaFields = []string{
		"url", "height", "width", "media_key", "variants",
	}
)

type Client struct {
	httpClient *http.Client
	token      string

	logger *util.Logger
}

func NewClient(token string) *Client {
	return &Client{
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
			if bitrate < v.Bitrate {
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

func (c *Client) GetTweets(ids []string) ([]*Tweet, error) {
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
	request, err := http.NewRequest("GET", url.String(), nil)
	if err != nil {
		return nil, err
	}
	request.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.token))
	request.Header.Set("Content-Type", "application/json")

	response, err := c.httpClient.Do(request)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	bytes, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}
	result := &Response{}
	if err = json.Unmarshal(bytes, result); err != nil {
		return nil, err
	}
	var mediaByKey map[string][]Media
	if result.Includes != nil {
		mediaByKey = util.GroupBy(result.Includes.Media, func(m Media) string {
			return m.MediaKey
		})
	} else {
		mediaByKey = map[string][]Media{}
	}
	for _, tweet := range result.Data {
		for _, mk := range tweet.Attachments.MediaKeys {
			if medias, ok := mediaByKey[mk]; ok {
				tweet.Media = append(tweet.Media, getBestQualityMedia(medias))
			}
		}
	}
	return result.Data, nil
}
