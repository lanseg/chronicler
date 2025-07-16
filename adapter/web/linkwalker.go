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

type VisitRule = func(*LinkWalker, *url.URL, *url.URL) bool

func IsSameHost(lw *LinkWalker, parent *url.URL, link *url.URL) bool {
	return (lw.Root == nil || common.IsSameHost(lw.Root, parent)) && common.IsSameHost(parent, link)
}

func IsHTTP(lw *LinkWalker, parent *url.URL, link *url.URL) bool {
	return link.Scheme == "http" || link.Scheme == "https"
}

func IsHTML(lw *LinkWalker, parent *url.URL, link *url.URL) bool {
	mime := common.GuessMimeType(link.String())
	return mime == "" || strings.HasPrefix(mime, "text/html")
}

type LinkWalker struct {
	logger *common.Logger

	Rules    []VisitRule
	Root     *url.URL        `json:"root"`
	MaxLinks int             `json:"max_links"`
	Links    map[string]bool `json:"links"`
}

func NewWalker(root *url.URL, maxLinks int) *LinkWalker {
	return &LinkWalker{
		logger: common.NewLogger("LinkWalker"),
		Root:   root,
		Rules: []VisitRule{
			IsHTTP,
			IsSameHost,
			IsHTML,
		},
		MaxLinks: maxLinks,
		Links:    map[string]bool{root.String(): false},
	}
}

func (lw *LinkWalker) shouldVisit(parent *url.URL, link *url.URL) bool {
	_, knownLink := lw.Links[link.String()]
	if knownLink || lw.MaxLinks <= len(lw.Links) {
		return false
	}
	for _, rule := range lw.Rules {
		if !rule(lw, parent, link) {
			return false
		}
	}
	return true
}

func (lw *LinkWalker) MarkVisited(links []string) {
	for _, l := range links {
		lw.Links[l] = true
	}
}

func (lw *LinkWalker) NextToVisit(count int) []string {
	if count == 0 {
		return []string{}
	}
	result := []string{}
	for link, known := range lw.Links {
		if known {
			continue
		}
		result = append(result, link)
		count--
		if count == 0 {
			break
		}
	}
	return result
}

func (lw *LinkWalker) AddToVisit(link string) {
	if _, ok := lw.Links[link]; !ok {
		lw.Links[link] = false
	}
}

func (lw *LinkWalker) FindLinks(baseUrl *url.URL, data []byte) map[string](map[string]bool) {
	allLinks := map[string](map[string]bool){}
	reader := parser.NewHtmlReader(bytes.NewReader(data))
	for reader.NextToken() {
		for attr := range linkAttr {
			if href, ok := reader.Attr(attr); ok && href != "" {
				if strings.HasPrefix(href, "#") {
					continue
				}
				h, err := common.ParseUrlDefaults(href, baseUrl)
				if err != nil {
					lw.logger.Errorf("cannot parse attribute %q (%q) from token %q as url: %s",
						attr, href, reader.Raw(), err)
					continue
				}
				actualLink := common.NormalizeURL(h).String()
				if allLinks[actualLink] == nil {
					allLinks[actualLink] = map[string]bool{}
				}
				if href != actualLink {
					allLinks[actualLink][href] = true
				}
			}
		}
		for _, href := range linkRe.FindAllString(reader.Raw(), -1) {
			h, err := common.ParseUrlDefaults(href, baseUrl)
			if err != nil {
				lw.logger.Errorf("cannot parse %q from token %q as url: %s", href, reader.Raw(), err)
				continue
			}
			actualLink := common.NormalizeURL(h).String()
			if allLinks[actualLink] == nil {
				allLinks[actualLink] = map[string]bool{}
			}
			if href != actualLink {
				allLinks[actualLink][href] = true
			}
		}
	}
	for k := range allLinks {
		linkAsUrl, err := url.Parse(k)
		if err != nil {
			continue
		}
		linkAsUrl.Fragment = ""
		if lw.shouldVisit(baseUrl, linkAsUrl) {
			lw.Links[linkAsUrl.String()] = false
		}
	}
	return allLinks
}
