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

type webAdapter struct {
	adapter.Adapter

	logger *common.Logger
	client *http.Client
}

func NewAdapter(client *http.Client) adapter.Adapter {
	return &webAdapter{
		client: client,
		logger: common.NewLogger("WebAdapter"),
	}
}

func (wa *webAdapter) Match(link *opb.Link) bool {
	return false
	// u, err := url.Parse(link.Href)
	// if err != nil {
	// 	wa.logger.Warningf("Not matching, link %s is not an url:%s ", link, err)
	// 	return false
	// }
	// matches := u.Scheme == "http" || u.Scheme == "https"
	// if !matches {
	// 	wa.logger.Warningf("Not a http/https link: Scheme is %q", u.Scheme)
	// }
	//return matches
}

func (pa *webAdapter) Get(link *opb.Link) ([]*opb.Object, error) {
	resp, err := pa.client.Get(link.Href)
	if err != nil {
		return nil, err
	}

	actualUrl := resp.Request.URL
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	re := regexp.MustCompile("(http|ftp|https):\\/\\/([\\w_-]+(?:(?:\\.[\\w_-]+)+))([\\w.,@?^=%&:\\/~+#-]*[\\w@?^=%&\\/~+#-])")
	urls := map[string]bool{}
	reader := parser.NewHtmlReader(bytes.NewReader(data))
	for !reader.NextToken() {
		if href := reader.Attr("href"); href != "" {
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
		for _, u := range re.FindAllString(reader.Raw(), -1) {
			urls[u] = true
		}
	}

	attachments := []*opb.Attachment{}
	for u := range urls {
		attachments = append(attachments, &opb.Attachment{
			Url:  u,
			Mime: mime.TypeByExtension(filepath.Ext(u)),
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
