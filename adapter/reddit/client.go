package reddit

import (
	"chronicler/adapter"
	"chronicler/common"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
)

var (
	redditRe = regexp.MustCompile(`reddit.com/r/(?P<subreddit>[^/]*)/comments/(?P<postId>[^/]*)/?(?P<maybePostName>[^/$]*)(?:/(?P<maybeCommentId>[^/$]*))?`)
)

type RedditPostDef struct {
	Subreddit      string
	PostId         string
	MaybePost      string
	MaybeCommentId string
}

type GetPostResponse struct {
	Entities []*Entity `json:"entities"`
	More     []string  `json:"more"`
}

func ParseLink(link string) *RedditPostDef {
	subexp := redditRe.SubexpNames()
	submatch := redditRe.FindStringSubmatch(link)
	result := &RedditPostDef{}
	for i, key := range submatch {
		switch subexp[i] {
		case "subreddit":
			result.Subreddit = key
		case "postId":
			result.PostId = key
		case "maybePostName":
			result.MaybePost = key
		case "maybeCommentId":
			result.MaybeCommentId = key
		}
	}
	return result
}

type Client interface {
	GetPost(def *RedditPostDef) ([]*Entity, error)
}

type RedditAuth struct {
	AccessToken  string
	RefreshToken string
}

type redditClient struct {
	auth       *RedditAuth
	httpClient adapter.HttpClient
	logger     *common.Logger
}

func (rc *redditClient) get(def *RedditPostDef) ([]Thing[Listing[Thing[Entity]]], error) {
	request := &http.Request{
		Method: "GET",
		URL: &url.URL{
			Scheme:   "https",
			Host:     "www.reddit.com",
			Path:     fmt.Sprintf("/r/%s/comments/%s.json", def.Subreddit, def.PostId),
			RawQuery: "threaded=false&limit=1000000",
		},
		Header: http.Header{
			"Content-type": []string{"application/x-www-form-urlencoded"},
		},
	}
	if rc.auth != nil {
		request.Header.Add("Authentication", fmt.Sprintf("Bearer %s", rc.auth.AccessToken))
	}
	resp, err := rc.httpClient.Do(request)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	responseBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	result := []Thing[Listing[Thing[Entity]]]{}
	if err = json.Unmarshal(responseBytes, &result); err != nil {
		return nil, err
	}
	return result, nil
}

func (rc *redditClient) fetch(def *RedditPostDef) (*GetPostResponse, error) {
	result := &GetPostResponse{}
	redditPost, err := rc.get(def)
	if err != nil {
		return nil, err
	}
	toFetch := map[string]bool{}
	for _, th := range redditPost {
		for _, the := range th.Data.Children {
			if the.Kind == "more" {
				toFetch[the.Data.Id] = true
				for _, ch := range the.Data.Children {
					toFetch[ch] = true
				}
			} else {
				result.Entities = append(result.Entities, &the.Data)
			}
		}
		th.Data.Children = nil
	}

	for k := range toFetch {
		result.More = append(result.More, k)
	}
	return result, nil
}

func (rc *redditClient) GetPost(def *RedditPostDef) ([]*Entity, error) {
	rc.logger.Infof("Loading post and comments from %s/%s", def.Subreddit, def.MaybePost)
	post, err := rc.fetch(def)
	if err != nil {
		return nil, err
	}
	rc.logger.Infof("Loaded entities from %s/%s: %d, still to load at least %d",
		def.Subreddit, def.MaybePost, len(post.Entities), len(post.More))
	return post.Entities, nil
}

func NewAnonymousClient(client adapter.HttpClient) Client {
	return NewClient(client, nil)
}

func NewClient(client adapter.HttpClient, auth *RedditAuth) Client {
	return &redditClient{
		auth:       auth,
		logger:     common.NewLogger("RedditClient"),
		httpClient: client,
	}
}
