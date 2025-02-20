package web

import (
	"bytes"
	"net/url"

	"chronicler/parser"
)

func FindLinks(baseUrl *url.URL, data []byte) map[string]bool {
	urls := map[string]bool{}
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
					urls[h.String()] = true
				}
			}
		}
		for _, u := range linkRe.FindAllString(reader.Raw(), -1) {
			urls[u] = true
		}
	}
	return urls
}
