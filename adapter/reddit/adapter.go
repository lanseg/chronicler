package reddit

import (
	"fmt"
	"mime"
	"net/url"
	"path/filepath"
	"strings"

	"chronicler/adapter"
	"chronicler/common"
	opb "chronicler/proto"
)

type redditAdapter struct {
	adapter.Adapter

	logger *common.Logger
	client Client
}

func getMime(href string) string {
	fileName := ""
	if u, err := url.Parse(href); err == nil {
		fileName = u.Path
	} else {
		fileName = href
	}
	return mime.TypeByExtension(filepath.Ext(fileName))
}

func NewAnonymousAdapter(client adapter.HttpClient) adapter.Adapter {
	return &redditAdapter{
		logger: common.NewLogger("RedditAdapter"),
		client: NewAnonymousClient(client),
	}
}

func NewAdapter(client adapter.HttpClient, auth *RedditAuth) adapter.Adapter {
	return &redditAdapter{
		logger: common.NewLogger("RedditAdapter"),
		client: NewClient(client, auth),
	}
}

func (ta *redditAdapter) Match(link *opb.Link) bool {
	postDef := ParseLink(link.Href)
	return postDef.PostId != "" && postDef.Subreddit != ""
}

func (ta *redditAdapter) Get(link *opb.Link) ([]*opb.Object, error) {
	postDef := ParseLink(link.Href)
	if postDef.Subreddit == "" || postDef.PostId == "" {
		return nil, fmt.Errorf("%s is not a Reddit post link", link.Href)
	}
	postData, err := ta.client.GetPost(postDef)
	if err != nil {
		return nil, err
	}
	ta.logger.Infof("Loaded entities from %s/%s: %d, still to load at least %d",
		postDef.Subreddit, postDef.PostId, len(postData.Entities), len(postData.More))
	entities := postData.Entities

	toload := postData.More
	batchSize := 200
	if len(toload) > 0 {
		for start := 0; start < len(toload); start += batchSize {
			end := start + batchSize
			if end > len(toload) {
				end = len(toload)
			}
			ta.logger.Infof("Loading children for %s, [%04d of %04d]", postDef.PostId, end, len(toload))
			resp, err := ta.client.GetChildren(postDef, toload[start:end])
			if err != nil {
				ta.logger.Warningf("Failed while loading children: %s", err)
				break
			}
			entities = append(entities, resp.Entities...)
		}
	}
	result := []*opb.Object{}
	for _, e := range entities {
		links := map[string]bool{}
		if e.Media != nil && e.Media.RedditVideo != nil {
			links[e.Media.RedditVideo.FallbackUrl] = true
		}
		if e.SecureMedia != nil && e.SecureMedia.RedditVideo != nil {
			links[e.SecureMedia.RedditVideo.FallbackUrl] = true
		}
		if e.Preview != nil {
			for _, img := range e.Preview.Images {
				links[img.Source.Url] = true
			}
		}
		for _, v := range e.MediaMetadata {
			if v.Source.Url != "" {
				links[v.Source.Url] = true
			}
			if v.Source.Mp4 != "" {
				links[v.Source.Mp4] = true
			}
			if v.Source.Gif != "" {
				links[v.Source.Gif] = true
			}
		}
		attachments := []*opb.Attachment{}
		for l := range links {
			attachments = append(attachments, &opb.Attachment{
				Url:  strings.ReplaceAll(l, "&amp;", "&"),
				Mime: getMime(l),
			})

		}

		content := []*opb.Content{}
		if e.BodyHtml != "" {
			content = append(content, &opb.Content{Text: e.BodyHtml, Mime: "text/html"})
		} else if e.Body != "" {
			content = append(content, &opb.Content{Text: e.BodyHtml, Mime: "text/plain"})
		}
		stats := []*opb.Stats{}
		if e.Ups > 0 {
			stats = append(stats, &opb.Stats{Type: opb.Stats_UPVOTE, Counter: int64(e.Ups)})
		}
		if e.Downs > 0 {
			stats = append(stats, &opb.Stats{Type: opb.Stats_DOWNVOTE, Counter: int64(e.Downs)})
		}
		authorIdName := e.AuthorFullName
		parent := e.ParentId
		if len(parent) > 3 {
			parent = parent[3:]
		}
		if len(authorIdName) > 3 {
			authorIdName = authorIdName[3:]
		}
		result = append(result, &opb.Object{
			Id:     e.Id,
			Parent: parent,
			CreatedAt: &opb.Timestamp{
				Seconds: int64(e.CreatedUtc),
			},
			Stats:      stats,
			Attachment: attachments,
			Generator:  []*opb.Generator{{Id: authorIdName, Name: e.Author}},
			Content:    content,
		})
	}
	ta.logger.Infof("Loaded objects for post %s: %d", postDef.PostId, len(result))
	return result, nil
}
