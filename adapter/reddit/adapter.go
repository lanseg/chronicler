package reddit

import (
	"fmt"
	"mime"
	"net/url"
	"path/filepath"

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

func NewAdapter(client adapter.HttpClient) adapter.Adapter {
	return &redditAdapter{
		logger: common.NewLogger("TwitterAdapter"),
		client: NewAnonymousClient(client),
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
	entities, err := ta.client.GetPost(postDef)
	if err != nil {
		return nil, err
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
			links[v.Original.Url] = true
		}
		attachments := []*opb.Attachment{}
		for l := range links {
			attachments = append(attachments, &opb.Attachment{
				Url:  l,
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
	return result, nil
}
