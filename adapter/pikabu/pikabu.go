package pikabu

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"time"

	"github.com/lanseg/golang-commons/collections"
	cm "github.com/lanseg/golang-commons/common"
	"golang.org/x/text/encoding/charmap"

	"chronicler/adapter"
	rpb "chronicler/records/proto"
	"chronicler/webdriver"
)

const (
	pikabuStoryRe = "/story/.*_(?P<story_id>[0-9]+)[#/]?"
)

func matchLink(link string) *rpb.Source {
	if link == "" {
		return nil
	}
	u, err := url.Parse(link)
	if err != nil || u.Host != "pikabu.ru" {
		return nil
	}
	linkMatcher := regexp.MustCompile(pikabuStoryRe)
	matches := collections.NewMap(
		linkMatcher.SubexpNames(),
		linkMatcher.FindStringSubmatch(u.Path))
	if match, ok := matches["story_id"]; ok && match != "" {
		return &rpb.Source{
			ChannelId: match,
			Url:       u.String(),
			Type:      rpb.SourceType_PIKABU,
		}
	}
	return nil
}

type pikabuAdapter struct {
	adapter.Adapter

	linkMatcher *regexp.Regexp
	logger      *cm.Logger
	browser     webdriver.Browser
}

func NewPikabuAdapter(browser webdriver.Browser) adapter.Adapter {
	return &pikabuAdapter{
		logger:  cm.NewLogger("PikabuAdapter"),
		browser: browser,
	}
}

func (p *pikabuAdapter) FindSources(r *rpb.Record) []*rpb.Source {
	result := []*rpb.Source{}
	for _, link := range r.Links {
		if src := matchLink(link); src != nil {
			result = append(result, src)
		}
	}
	return result
}

func (p *pikabuAdapter) getContentWebdriver(storyId string) string {
	content := ""
	p.browser.RunSession(func(w webdriver.WebDriver) {
		w.Navigate(fmt.Sprintf("https://pikabu.ru/story/_%s", storyId))
		w.GetPageSource().IfPresent(func(s string) {
			content = s
		})
	})
	return content
}

func (p *pikabuAdapter) getContentHttpPlain(storyId string) string {
	response, err := http.Get(fmt.Sprintf("https://pikabu.ru/story/_%s", storyId))
	if err != nil {
		p.logger.Warningf("Error while making plain http request: %s", err)
		return ""
	}
	data, err := io.ReadAll(response.Body)
	if err != nil {
		p.logger.Warningf("Error while reading http data: %s", err)
		return ""
	}
	dataUtf, err := charmap.Windows1251.NewDecoder().Bytes(data)
	if err != nil {
		p.logger.Warningf("Error while decoding http data: %s", err)
		return ""
	}
	return string(dataUtf)
}

func (p *pikabuAdapter) GetResponse(rq *rpb.Request) []*rpb.Response {
	p.logger.Debugf("Got new request: %s", rq)

	content := ""
	if rq.Config != nil && rq.Config.Engine == rpb.WebEngine_HTTP_PLAIN {
		p.logger.Debugf("Using plain http client to get the content")
		content = p.getContentHttpPlain(rq.Target.ChannelId)
	} else {
		p.logger.Debugf("Using webdriver to get the content")
		content = p.getContentWebdriver(rq.Target.ChannelId)
	}

	p.logger.Infof("Loaded page content, got string of %d: %s", len(content), cm.Ellipsis(content, 100, true))
	resp, err := parsePost(content, time.Now)
	if resp == nil || err != nil {
		p.logger.Warningf("Error while parsing request: %s", rq)
		return []*rpb.Response{}
	}
	resp.Request = rq
	resp.Result[0].Records[0].Source.ChannelId = rq.Target.ChannelId
	resp.Result[0].Id = cm.UUID4For(rq.Target.ChannelId)
	p.logger.Infof("Parsed post with %d record(s) and %d user(s)",
		len(resp.Result[0].Records), len(resp.Result[0].UserMetadata))
	return []*rpb.Response{resp}
}

func (p *pikabuAdapter) SendMessage(m *rpb.Message) {
}
