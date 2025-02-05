package fourchan

import (
	"fmt"
	"mime"
	"net/url"
	"regexp"
	"strings"

	"chronicler/adapter"
	"chronicler/common"

	opb "chronicler/proto"
)

var (
	threadId = regexp.MustCompile(`https?:\/\/boards.4chan.org\/([a-z]+)\/thread\/(\d+)(\#p(\d+))?`)
)

type FourChanLinkDef struct {
	Board    string
	ThreadId string
	PostId   string
}

func ParseLink(link string) *FourChanLinkDef {
	maybeId := threadId.FindAllStringSubmatch(link, 1)
	if len(maybeId) == 0 {
		return nil
	}
	return &FourChanLinkDef{
		Board:    maybeId[0][1],
		ThreadId: maybeId[0][2],
		PostId:   maybeId[0][3],
	}
}

type fourchanAdapter struct {
	adapter.Adapter

	logger     *common.Logger
	httpClient adapter.HttpClient
}

func NewAdapter(client adapter.HttpClient) adapter.Adapter {
	return &fourchanAdapter{
		httpClient: client,
		logger:     common.NewLogger("FourChanAdapter"),
	}
}

func (fca *fourchanAdapter) Match(link *opb.Link) bool {
	u, err := url.Parse(link.Href)
	if err != nil {
		fca.logger.Warningf("Not matching, link %s is not an url:%s ", link, err)
		return false
	}
	matches := u.Scheme == "http" || u.Scheme == "https"
	if !matches {
		fca.logger.Warningf("Not a http/https link: Scheme is %q", u.Scheme)
	}
	return ParseLink(link.Href) != nil
}

func (fca *fourchanAdapter) Get(link *opb.Link) ([]*opb.Object, error) {
	post := ParseLink(link.Href)
	if post == nil {
		return nil, fmt.Errorf("link %q is not a 4chan post", link)
	}
	fca.logger.Infof("Loading 4chan post %q", post)
	posts, err := GetThread(fca.httpClient, post.Board, post.ThreadId)
	if err != nil {
		return nil, err
	}

	fca.logger.Debugf("Thread contains %d post(s)", len(posts))
	result := []*opb.Object{}
	for _, p := range posts {
		obj := &opb.Object{
			Id: fmt.Sprintf("%d", p.No),
			CreatedAt: &opb.Timestamp{
				Seconds: p.Time,
			},
			Generator: []*opb.Generator{{
				Name: p.Name,
				Id:   p.Id,
			}},
			Content: []*opb.Content{{Text: p.Com, Mime: "text/html"}},
		}
		if p.Resto != 0 {
			obj.Parent = fmt.Sprintf("%d", p.Resto)
		}

		parentPrefix := `<a href="#p`
		if strings.HasPrefix(p.Com, parentPrefix) {
			from := len(parentPrefix)
			obj.Parent = p.Com[from : from+strings.IndexRune(p.Com[from:], '"')]
		}
		if p.Sub != "" {
			obj.Content = append([]*opb.Content{{Text: p.Sub, Mime: "text/html"}}, obj.Content...)
		}
		if p.MD5 != "" {
			obj.Attachment = append(obj.Attachment, &opb.Attachment{
				Checksum: fmt.Sprintf("md5:%s", p.MD5),
				Url:      fmt.Sprintf("https://i.4cdn.org/%s/%d%s", post.Board, p.Tim, p.Ext),
				Mime:     mime.TypeByExtension(p.Ext),
			})
		}

		result = append(result, obj)
	}
	return result, nil
}
