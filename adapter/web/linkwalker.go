package web

import (
	"bytes"
	"net/url"

	"chronicler/common"
	"chronicler/parser"
)

type LinkWalker struct {
	Visited map[string]bool `json:"visited"`
	ToVisit map[string]bool `json:"tovisit"`
}

func (lw *LinkWalker) MarkVisited(links []string) {
	for _, l := range links {
		delete(lw.ToVisit, l)
		lw.Visited[l] = true
	}
}

func (lw *LinkWalker) NextToVisit(count int) []string {
	result := []string{}
	added := 0
	for k := range lw.ToVisit {
		result = append(result, k)
		added++
		if added == count {
			break
		}
	}
	return result
}

func (lw *LinkWalker) AddToVisit(link string) {
	lw.ToVisit[link] = true
}

func (lw *LinkWalker) shouldVisit(parent *url.URL, link *url.URL) bool {
	href := link.String()
	return !lw.Visited[href] && !lw.ToVisit[href] &&
		common.GuessMimeType(href) == "" &&
		(link.Scheme == "http" || link.Scheme == "https") && link.Hostname() == parent.Hostname()
}

func (lw *LinkWalker) FindLinks(baseUrl *url.URL, data []byte) map[string]bool {
	allLinks := map[string]bool{}
	reader := parser.NewHtmlReader(bytes.NewReader(data))
	for reader.NextToken() {
		for attr := range linkAttr {
			if href, ok := reader.Attr(attr); ok && href != "" {
				h, err := url.Parse(href)
				if err == nil {
					if h.Scheme == "" {
						h.Scheme = baseUrl.Scheme
					}
					if h.Host == "" {
						h.Host = baseUrl.Host
					}
					if h.Path == "" {
						h.Path = baseUrl.Path
					}
					allLinks[h.String()] = true
				}
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
			lw.ToVisit[k] = true
		}
	}
	return allLinks
}
