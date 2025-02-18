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
	"strings"
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

type GetMoreChildrenResponse struct {
	Json struct {
		Errors []interface{} `json:"errors"`
		Data   struct {
			Things []*Thing[Entity] `json:"things"`
		} `json:"data"`
	} `json:"json"`
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

func parsePostResponse(redditPost []Thing[Listing[Thing[Entity]]]) *GetPostResponse {
	result := &GetPostResponse{}
	more := map[string]bool{}
	for _, th := range redditPost {
		for _, the := range th.Data.Children {
			if the.Kind == "more" {
				for _, ch := range the.Data.Children {
					more[ch] = true
				}
			} else {
				result.Entities = append(result.Entities, &the.Data)
			}
		}
		th.Data.Children = nil
	}
	for k := range more {
		result.More = append(result.More, k)
	}
	return result
}

type Client interface {
	GetPost(def *RedditPostDef) (*GetPostResponse, error)
	GetChildren(def *RedditPostDef, childIds []string) (*GetPostResponse, error)
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

func (rc *redditClient) performRequest(linkStr string, auth bool) ([]byte, error) {
	link, err := url.Parse(linkStr)
	if err != nil {
		return nil, err
	}
	request := &http.Request{
		Method: "GET",
		URL:    link,
		Header: http.Header{
			"Accept":     []string{"application/json"},
			"User-Agent": []string{"Mozilla/5.0 (X11; Linux x86_64; rv:58.0) Gecko/20100101 Firefox/58.0"},
		},
	}
	if auth && rc.auth != nil {
		request.Header.Add("Authorization", fmt.Sprintf("Bearer %s", rc.auth.AccessToken))
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
	return responseBytes, nil
}

func (rc *redditClient) GetPost(def *RedditPostDef) (*GetPostResponse, error) {
	redditPostBytes, err := rc.performRequest(
		fmt.Sprintf("https://www.reddit.com/r/%s/comments/%s.json?threaded=false&limit=1000000",
			def.Subreddit, def.PostId), false)
	if err != nil {
		return nil, err
	}
	result := []Thing[Listing[Thing[Entity]]]{}
	if err = json.Unmarshal(redditPostBytes, &result); err != nil {
		return nil, err
	}
	return parsePostResponse(result), nil
}

func (rc *redditClient) GetChildren(rp *RedditPostDef, childIds []string) (*GetPostResponse, error) {
	redditPostBytes, err := rc.performRequest(fmt.Sprintf(
		"https://oauth.reddit.com/api/morechildren?api_type=json&link_id=t3_%s&children=%s&limit_children=true&sort=new",
		rp.PostId, strings.Join(childIds, ",")), true)
	if err != nil {
		return nil, err
	}
	result := &GetMoreChildrenResponse{}
	if err = json.Unmarshal(redditPostBytes, &result); err != nil {
		return nil, err
	}
	entities := []*Entity{}
	toLoad := []string{}
	for _, th := range result.Json.Data.Things {
		if th.Kind == "more" {
			toLoad = append(toLoad, th.Data.Children...)
		}
		entities = append(entities, &th.Data)
	}
	return &GetPostResponse{
		Entities: entities,
		More:     toLoad,
	}, nil
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
