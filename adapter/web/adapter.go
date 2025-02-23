package web

import (
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"sort"
	"time"

	"chronicler/adapter"
	"chronicler/common"
	opb "chronicler/proto"
)

var (
	linkRe   = regexp.MustCompile(`(http|ftp|https)://([\w_-]+(?:(?:\.[\w_-]+)+))([\w.,@?^=%&:\/~+#-]*[\w@?^=%&\/~+#-])`)
	linkAttr = map[string]bool{
		"href":       true,
		"src":        true,
		"background": true,
		"profile":    true,
		"longdesc":   true,
		"icon":       true,
		"manifest":   true,
		"poster":     true,
		"data-src":   true,
	}
)

type webAdapter struct {
	adapter.Adapter

	walker *LinkWalker
	logger *common.Logger
	client adapter.HttpClient
}

func NewAdapter(client adapter.HttpClient) adapter.Adapter {
	walker := &LinkWalker{Visited: map[string]bool{}, ToVisit: map[string]bool{}}
	data, err := os.ReadFile("walker.json")
	if err != nil {
		json.Unmarshal(data, walker)
	}
	return &webAdapter{
		client: client,
		walker: walker,
		logger: common.NewLogger("WebAdapter"),
	}
}

func (wa *webAdapter) saveWalker() error {
	data, err := json.Marshal(wa.walker)
	if err != nil {
		return err
	}
	return os.WriteFile("walker.json", data, 0777)
}

func (wa *webAdapter) Match(link *opb.Link) bool {
	u, err := url.Parse(link.Href)
	if err != nil {
		wa.logger.Warningf("Not matching, link %s is not an url:%s ", link, err)
		return false
	}
	matches := u.Scheme == "http" || u.Scheme == "https"
	if !matches {
		wa.logger.Warningf("Not a http/https link: Scheme is %q", u.Scheme)
	}
	return matches
}

func (wa *webAdapter) Get(link *opb.Link) ([]*opb.Object, error) {
	wa.walker.AddToVisit(link.Href)
	result := []*opb.Object{}
	maxLinks := 10000000
	for i := 0; i < maxLinks; i++ {
		next := wa.walker.NextToVisit(1)
		if len(next) == 0 {
			break
		}
		wa.walker.MarkVisited(next)

		current := next[0]
		wa.logger.Infof("Resolving page [%d of %d (%d)]: %s", i, len(wa.walker.ToVisit), maxLinks, current)
		url, err := url.Parse(current)
		if err != nil {
			continue
		}
		resp, err := wa.client.Do(&http.Request{Method: "GET", URL: url})
		if err != nil {
			return nil, err
		}

		data, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}

		attachments := []*opb.Attachment{}
		for u := range wa.walker.FindLinks(resp.Request.URL, data) {
			mime := common.GuessMimeType(u)
			attachments = append(attachments, &opb.Attachment{
				Url:  u,
				Mime: mime,
			})
		}
		sort.Slice(attachments, func(i, j int) bool {
			return attachments[i].Url < attachments[j].Url
		})

		result = append(result, &opb.Object{
			Id: link.Href,
			Content: []*opb.Content{
				{
					Text: string(data),
					Mime: "text/html",
				},
			},
			Attachment: attachments,
		})
		if (i % 100) == 0 {
			if err = wa.saveWalker(); err != nil {
				wa.logger.Warningf("Cannot save link walker status: %s", err)
			}
		}
		time.Sleep(200 * time.Microsecond)
	}
	return result, nil
}
