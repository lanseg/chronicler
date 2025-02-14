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
		attachments := []*opb.Attachment{}
		if e.Media != nil && e.Media.RedditVideo != nil {
			attachments = append(attachments, &opb.Attachment{
				Url:  e.Media.RedditVideo.FallbackUrl,
				Mime: getMime(e.Media.RedditVideo.FallbackUrl),
			})
		}
		if e.Preview != nil {
			for _, img := range e.Preview.Images {
				attachments = append(attachments, &opb.Attachment{
					Url:  img.Source.Url,
					Mime: getMime(img.Source.Url),
				})
			}
		}
		for _, v := range e.MediaMetadata {
			attachments = append(attachments, &opb.Attachment{
				Url:  v.Original.Url,
				Mime: v.MimeType,
			})
		}
		content := []*opb.Content{}
		if e.BodyHtml != "" {
			content = append(content, &opb.Content{Text: e.BodyHtml, Mime: "text/html"})
		} else if e.Body != "" {
			content = append(content, &opb.Content{Text: e.BodyHtml, Mime: "text/plain"})
		}
		result = append(result, &opb.Object{
			Id:     e.Id,
			Parent: e.ParentId,
			CreatedAt: &opb.Timestamp{
				Seconds: int64(e.CreatedUtc),
			},
			Stats: []*opb.Stats{
				{Type: opb.Stats_UPVOTE, Counter: int64(e.Ups)},
				{Type: opb.Stats_DOWNVOTE, Counter: int64(e.Downs)},
			},
			Attachment: attachments,
			Generator:  []*opb.Generator{{Id: e.AuthotFullName, Name: e.Author}},
			Content:    content,
		})
	}
	return result, nil
}
