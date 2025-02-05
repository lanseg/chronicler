package pikabu

import (
	"bytes"
	"fmt"
	"net/url"
	"regexp"
	"sort"
	"strings"

	"chronicler/adapter"
	"chronicler/common"
	opb "chronicler/proto"

	"golang.org/x/net/html"
)

var (
	storyId = regexp.MustCompile("pikabu.ru/story/[^/?]*_([0-9]+)")
)

type pikabuAdapter struct {
	adapter.Adapter

	logger *common.Logger
	client *Client
}

func NewAdapter(client adapter.HttpClient) adapter.Adapter {
	return &pikabuAdapter{
		client: NewClient(client),
		logger: common.NewLogger("PikabuAdapter"),
	}
}

func (pa *pikabuAdapter) getPostId(link *opb.Link) string {
	maybeId := storyId.FindAllStringSubmatch(link.Href, 1)
	if len(maybeId) == 0 {
		return ""
	}
	return maybeId[0][1]
}

func (pa *pikabuAdapter) Match(link *opb.Link) bool {
	_, err := url.Parse(link.Href)
	if err != nil {
		pa.logger.Debugf("Invalid link %s:%s ", link, err)
		return false
	}
	id := pa.getPostId(link)
	if id == "" {
		pa.logger.Debugf("Doesn't look like a pikabu post link:%s ", link)
		return false
	}
	return true
}

func (pa *pikabuAdapter) Get(link *opb.Link) ([]*opb.Object, error) {
	id := pa.getPostId(link)
	if id == "" {
		return nil, fmt.Errorf("no post id in the link: %s", link.Href)
	}
	pa.logger.Debugf("Loading post %s", id)
	postText, err := pa.client.GetPost(id)
	if err != nil {
		return nil, err
	}
	objs, err := NewPikabuParser(bytes.NewReader([]byte(postText))).Parse()
	if err != nil {
		return nil, err
	}
	objects := map[string]*opb.Object{}
	for _, post := range objs {
		objects[post.Id] = post
	}

	ids, _ := getCommentIds(postText)
	pa.logger.Debugf("Loading %d comments for post %s", len(ids), id)

	commText, _ := pa.client.GetComments(ids)
	for _, c := range commText {
		objs, _ := NewPikabuParser(bytes.NewReader([]byte(c.Html))).Parse()
		for _, obj := range objs {
			if obj.Parent == "0" || obj.Parent == "" {
				obj.Parent = id
			}
			objects[obj.Id] = obj
		}
	}

	pa.logger.Debugf("Loaded %d of %d comments for post %s", len(commText), len(ids), id)
	result := make([]*opb.Object, len(objects))
	i := 0
	for _, c := range objects {
		result[i] = c
		i++
	}
	sort.Slice(result, func(i, j int) bool {
		if result[i].CreatedAt.Seconds > result[j].CreatedAt.Seconds {
			return true
		} else if result[i].CreatedAt.Seconds < result[j].CreatedAt.Seconds {
			return false
		}
		return result[i].Id < result[j].Id
	})
	return result, nil
}

func getCommentIds(doc string) ([]string, error) {
	hr := html.NewTokenizer(bytes.NewReader([]byte(doc)))
	inTree := false
	for {
		if tokenType := hr.Next(); tokenType == html.ErrorToken {
			break
		}
		tok := hr.Token()
		if tok.Data == "script" && strings.Contains(string(hr.Raw()), "comments-tree") {
			inTree = true
			continue
		}
		if inTree {
			return ResolveCommentTree(string(hr.Raw()))
		}
	}
	return []string{}, nil
}
