package pikabu

import (
	"fmt"
	"net/url"
	"regexp"

	"chronicler/adapter"
	rpb "chronicler/records/proto"
	"chronicler/webdriver"

	"github.com/lanseg/golang-commons/collections"
	cm "github.com/lanseg/golang-commons/common"
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

func (p *pikabuAdapter) GetResponse(rq *rpb.Request) []*rpb.Response {
	p.logger.Infof("Got new request: %s", rq)

	content := ""
	p.browser.RunSession(func(w webdriver.WebDriver) {
		w.Navigate(fmt.Sprintf("https://pikabu.ru/story/_%s", rq.Target.ChannelId))
		w.GetPageSource().IfPresent(func(s string) {
			content = s
		})
	})

	p.logger.Infof("Loaded page content, got string of %d: %s", len(content), cm.Ellipsis(content, 100, true))
	resp := parsePost(content)
	resp.Request = rq
	resp.Result[0].Records[0].Source.ChannelId = rq.Target.ChannelId
	resp.Result[0].Id = cm.UUID4For(rq.Target)
	p.logger.Infof("Parsed post with %d record(s) and %d user(s)",
		len(resp.Result[0].Records), len(resp.Result[0].UserMetadata))
	return []*rpb.Response{resp}
}

func (p *pikabuAdapter) SendMessage(m *rpb.Message) {
}
