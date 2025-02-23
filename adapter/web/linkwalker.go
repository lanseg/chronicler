package web

import (
	"bytes"
	"net/url"

	"chronicler/common"
	"chronicler/parser"
)

type LinkWalker struct {
	visited map[string]bool
	toVisit map[string]bool
}

func (lw *LinkWalker) MarkVisited(links []string) {
	for _, l := range links {
		delete(lw.toVisit, l)
		lw.visited[l] = true
	}
}

func (lw *LinkWalker) NextToVisit(count int) []string {
	result := []string{}
	added := 0
	for k := range lw.toVisit {
		result = append(result, k)
		added++
		if added == count {
			break
		}
	}
	return result
}

func (lw *LinkWalker) AddToVisit(link string) {
	lw.toVisit[link] = true
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
		if !lw.visited[k] && !lw.toVisit[k] && common.GuessMimeType(k) == "" && (linkAsUrl.Scheme == "http" || linkAsUrl.Scheme == "https") && linkAsUrl.Hostname() == baseUrl.Hostname() {
			lw.toVisit[k] = true
		}
	}
	return allLinks
}
