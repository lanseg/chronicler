package pikabu

import (
	"bytes"
	"fmt"
	"net/http"
	"regexp"
	"sort"
	"strings"

	"chronicler/adapter"
	"chronicler/common"
	opb "chronicler/proto"

	"golang.org/x/net/html"
)

var (
	storyId = regexp.MustCompile("[0-9]+$")
)

type pikabuAdapter struct {
	adapter.Adapter

	logger *common.Logger
	client *Client
}

func NewAdapter(client *http.Client) adapter.Adapter {
	return &pikabuAdapter{
		client: NewClient(client),
		logger: common.NewLogger("PikabuAdapter"),
	}
}

func (pa *pikabuAdapter) Match(link *opb.Link) bool {
	// _, err := url.Parse(link.Href)
	// if err != nil {
	// 	pa.logger.Warningf("Not matching link %s:%s ", link, err)
	// 	return false
	// }
	// return true
	return false
}

func (pa *pikabuAdapter) Get(link *opb.Link) ([]*opb.Object, error) {
	id := storyId.FindString(link.Href)
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
		return result[i].CreatedAt.Seconds > result[j].CreatedAt.Seconds
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
