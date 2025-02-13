package web

import (
	"bytes"
	"io"
	"mime"
	"net/http"
	"net/url"
	"path/filepath"
	"regexp"
	"sort"

	"chronicler/adapter"
	"chronicler/common"
	"chronicler/parser"
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
	}
)

type webAdapter struct {
	adapter.Adapter

	logger *common.Logger
	client adapter.HttpClient
}

func NewAdapter(client adapter.HttpClient) adapter.Adapter {
	return &webAdapter{
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
	url, err := url.Parse(link.Href)
	if err != nil {
		return nil, err
	}
	resp, err := wa.client.Do(&http.Request{Method: "GET", URL: url})
	if err != nil {
		return nil, err
	}

	actualUrl := resp.Request.URL
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	urls := map[string]bool{}
	reader := parser.NewHtmlReader(bytes.NewReader(data))
	for reader.NextToken() {
		for attr := range linkAttr {
			if href, ok := reader.Attr(attr); ok && href != "" {
				h, err := url.Parse(href)
				if err == nil {
					if h.Scheme == "" {
						h.Scheme = actualUrl.Scheme
					}
					if h.Host == "" {
						h.Host = actualUrl.Host
					}
					if h.Path == "" {
						h.Path = actualUrl.Path
					}
					urls[h.String()] = true
				}
			}
		}
		for _, u := range linkRe.FindAllString(reader.Raw(), -1) {
			urls[u] = true
		}
	}

	attachments := []*opb.Attachment{}
	for u := range urls {
		queryPath := u
		url, err := url.Parse(u)
		if err == nil {
			queryPath = url.Path
		}
		attachments = append(attachments, &opb.Attachment{
			Url:  u,
			Mime: mime.TypeByExtension(filepath.Ext(queryPath)),
		})
	}
	sort.Slice(attachments, func(i, j int) bool {
		return attachments[i].Url < attachments[j].Url
	})
	return []*opb.Object{
		{
			Id: link.Href,
			Content: []*opb.Content{
				{
					Text: string(data),
					Mime: "text/html",
				},
			},
			Attachment: attachments,
		},
	}, nil
}
