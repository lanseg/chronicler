package web

import (
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"time"

	"chronicler/adapter"
	"chronicler/common"
	opb "chronicler/proto"
)

const (
	defaultDelay = 200 * time.Millisecond
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

	client adapter.HttpClient
	delay  time.Duration
	logger *common.Logger
}

func NewAdapter(client adapter.HttpClient) adapter.Adapter {
	return &webAdapter{
		delay:  defaultDelay,
		client: client,
		logger: common.NewLogger("WebAdapter"),
	}
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
	rootLink, err := url.Parse(link.Href)
	if err != nil {
		return nil, err
	}

	walker := NewWalker(rootLink)
	walkerData, err := os.ReadFile("walker.json")
	if err != nil {
		json.Unmarshal(walkerData, walker)
	}

	result := []*opb.Object{}
	for i := 0; ; i++ {
		next := walker.NextToVisit(1)
		if len(next) == 0 {
			break
		}
		walker.MarkVisited(next)

		current := next[0]
		wa.logger.Infof("Resolving page [%d of %d (%d)]: %s", i,
			len(walker.ToVisit), len(walker.ToVisit)+len(walker.Visited), current)
		url, err := url.Parse(current)
		if err != nil {
			wa.logger.Warningf("Ignoring invalid link %q: %s", current, err)
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
		for u := range walker.FindLinks(resp.Request.URL, data) {
			attachments = append(attachments, &opb.Attachment{
				Url:  u,
				Mime: common.GuessMimeType(u),
			})
		}

		result = append(result, &opb.Object{
			Id:         resp.Request.URL.String(),
			Attachment: attachments,
			Content:    []*opb.Content{{Text: string(data), Mime: "text/html"}},
		})
		if i%100 == 0 {
			walkerData, _ := json.Marshal(walker)
			os.WriteFile("walker.json", walkerData, 0777)
		}
		time.Sleep(wa.delay)
	}
	return result, nil
}
