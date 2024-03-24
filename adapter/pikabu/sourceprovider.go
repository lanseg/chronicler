package pikabu

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/lanseg/golang-commons/almosthtml"
	cm "github.com/lanseg/golang-commons/common"

	"chronicler/adapter"
	rpb "chronicler/records/proto"
)

type HttpClient interface {
	Get(string) (*http.Response, error)
}

type pikabuSourceProvider struct {
	adapter.SourceProvider

	skipSeen   bool
	seen       map[string]bool
	page       string
	httpClient HttpClient
	logger     *cm.Logger
}

func (p *pikabuSourceProvider) getPage() ([]byte, error) {
	resp, err := p.httpClient.Get(p.page)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return io.ReadAll(resp.Body)
}

func (p *pikabuSourceProvider) GetSources() []*rpb.Source {
	data, err := p.getPage()
	if err != nil {
		return []*rpb.Source{}
	}
	root, _ := almosthtml.ParseHTML(string(data))
	links := map[string]bool{}
	for _, article := range root.GetElementsByTags("article") {
		if article.Params["data-author-name"] == "specials" {
			continue
		}

		for _, n := range article.GetElementsByTags("a") {
			if href := n.GetAttribute("href").OrElse(""); strings.HasPrefix(href, "https://pikabu.ru/story") {
				url, err := url.Parse(href)
				if err != nil {
					p.logger.Warningf("Incorrect pikabu url %s: %s", href, err)
					continue
				}
				links[fmt.Sprintf("https://pikabu.ru/%s", url.Path)] = true
			}
		}
	}
	result := []*rpb.Source{}
	for link := range links {
		if p.skipSeen {
			if p.seen[link] {
				continue
			}
			p.seen[link] = true
		}
		result = append(result, matchLink(link))
	}
	return result
}

func NewDisputedProvider(client HttpClient) adapter.SourceProvider {
	return &pikabuSourceProvider{
		httpClient: client,
		page:       "https://pikabu.ru/disputed",
		logger:     cm.NewLogger("pikabu:disputed"),
	}
}

func NewFreshProvider(client HttpClient) adapter.SourceProvider {
	return &pikabuSourceProvider{
		httpClient: client,
		skipSeen:   true,
		seen:       map[string]bool{},
		page:       "https://pikabu.ru/new",
		logger:     cm.NewLogger("pikabu:new"),
	}
}

func NewHotProvider(client HttpClient) adapter.SourceProvider {
	return &pikabuSourceProvider{
		httpClient: client,
		page:       "https://pikabu.ru/hot",
		logger:     cm.NewLogger("pikabu:hot"),
	}
}
