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

func NewAdapter(client common.HttpClient) adapter.Adapter {
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
		return nil, fmt.Errorf("cannot get post %q: %w", id, err)
	}
	objs, err := NewPikabuParser(link.Href, bytes.NewReader([]byte(postText))).Parse()
	if err != nil {
		return nil, fmt.Errorf("cannot parse post %q: %w", id, err)
	}
	objects := map[string]*opb.Object{}
	for _, post := range objs {
		objects[post.Id] = post
	}

	ids, err := getCommentIds(postText)
	if err != nil {
		pa.logger.Warningf("Failed to fetch comment ids for post %s", id)
		ids = []string{}
	}
	pa.logger.Debugf("Loading %d comments for post %s", len(ids), id)

	commText, err := pa.client.GetComments(ids)
	if err != nil {
		pa.logger.Warningf("Failed to fetch comments by ids for post %s", id)
		commText = []*CommentData{}
	}
	for i, c := range commText {
		objs, err := NewPikabuParser(link.Href, bytes.NewReader([]byte(c.Html))).Parse()
		if err != nil {
			pa.logger.Warningf("Failed to parse comment %s/%s", link.Href, ids[i])
			continue
		}
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
	for tokenType := hr.Next(); !inTree && tokenType != html.ErrorToken; tokenType = hr.Next() {
		tok := hr.Token()
		if tok.Data == "script" && strings.Contains(string(hr.Raw()), "comments-tree") {
			inTree = true
		}
	}
	if inTree {
		return ResolveCommentTree(string(hr.Raw()))
	}
	return []string{}, nil
}
