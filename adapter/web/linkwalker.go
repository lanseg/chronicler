package web

import (
	"bytes"
	"net/url"
	"strings"

	"chronicler/common"
	"chronicler/parser"
)

const (
	defaultMaxLinks = 1000000
)

type LinkWalker struct {
	logger *common.Logger

	Root     *url.URL        `json:"root"`
	MaxLinks int             `json:"max_links"`
	Visited  map[string]bool `json:"visited"`
	ToVisit  map[string]bool `json:"to_visit"`
}

func NewWalker(root *url.URL) *LinkWalker {
	return &LinkWalker{
		logger:   common.NewLogger("LinkWalker"),
		Root:     root,
		MaxLinks: defaultMaxLinks,
		ToVisit:  map[string]bool{root.String(): true},
		Visited:  map[string]bool{},
	}
}

func (lw *LinkWalker) shouldVisit(parent *url.URL, link *url.URL) bool {
	href := link.String()
	mime := common.GuessMimeType(href)
	return lw.MaxLinks > (len(lw.ToVisit)+len(lw.Visited)) &&
		!lw.Visited[href] && !lw.ToVisit[href] &&
		(lw.Root == nil || common.IsSameHost(lw.Root, parent)) && common.IsSameHost(parent, link) &&
		(mime == "" || strings.HasPrefix(mime, "text/html")) &&
		(link.Scheme == "http" || link.Scheme == "https")
}

func (lw *LinkWalker) MarkVisited(links []string) {
	for _, l := range links {
		delete(lw.ToVisit, l)
		lw.Visited[l] = true
	}
}

func (lw *LinkWalker) NextToVisit(count int) []string {
	if count == 0 {
		return []string{}
	}
	result := []string{}
	for k := range lw.ToVisit {
		result = append(result, k)
		count--
		if count == 0 {
			break
		}
	}
	return result
}

func (lw *LinkWalker) AddToVisit(link string) {
	lw.ToVisit[link] = true
}

func (lw *LinkWalker) FindLinks(baseUrl *url.URL, data []byte) map[string]bool {
	allLinks := map[string]bool{}
	reader := parser.NewHtmlReader(bytes.NewReader(data))
	for reader.NextToken() {
		for attr := range linkAttr {
			if href, ok := reader.Attr(attr); ok && href != "" {
				if strings.HasPrefix(href, "#") {
					continue
				}
				h, err := common.ParseUrlDefaults(href, baseUrl)
				if err != nil {
					lw.logger.Errorf("cannot parse attribute %q from token %q as url: %s",
						attr, reader.Raw(), err)
				}
				allLinks[h.String()] = true
			}
		}
		for _, u := range linkRe.FindAllString(reader.Raw(), -1) {
			allLinks[u] = true
		}
	}
	for k := range allLinks {
		linkAsUrl, err := url.Parse(k)
		if err != nil {
			continue
		}
		linkAsUrl.Fragment = ""
		if lw.shouldVisit(baseUrl, linkAsUrl) {
			lw.ToVisit[linkAsUrl.String()] = true
		}
	}
	return allLinks
}
