package chronicler

import (
	rpb "chronicler/proto/records"
	"chronicler/util"
	"io"
	"net/http"
	"net/url"
	"strings"
	"web/htmlparser"
)

const (
	userAgent = "curl/7.54"
)

var (
    webpageFileTypes = util.NewSet([]string {
      "jpg", "png", "js", "css", "json", "ico",
    })
)

func fixLink(scheme string, host string, link string) string {
	u, err := url.Parse(link)
	if err != nil {
		return link
	}
	if u.Host == "" {
		u.Host = host
		u.Scheme = scheme
	}
	return u.String()
}

type Web struct {
	Chronicler

	name   string
	logger *util.Logger
	client *http.Client
}

func (w *Web) GetName() string {
	return w.name
}

func (w *Web) Get(link string) (*http.Response, error) {
	req, err := http.NewRequest("GET", link, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", userAgent)
	return w.client.Do(req)
}

func (w *Web) GetRecords(request *rpb.Request) (*rpb.RecordSet, error) {
	w.logger.Infof("Loading web page from %s", request.Source.Url)
	response, err := w.Get(request.Source.Url)
	if err != nil {
		return nil, err
	}
	requestUrl := response.Request.URL
	w.logger.Infof("Resolved actual URL as %s", requestUrl)
	defer response.Body.Close()
	body, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}
	record := &rpb.Record{
		Source: &rpb.Source{
			ChannelId: requestUrl.Host,
			Url:       requestUrl.String(),
			Type:      rpb.SourceType_WEB,
		},
		TextContent: string(body),
	}
	w.logger.Debugf("Parsing html content")
	root := htmlparser.ParseHtml(record.TextContent)
	linkNodes := root.GetElementsByTagNames("a", "img", "script", "link")
	w.logger.Debugf("Found %d external link(s)", len(linkNodes))
	for _, node := range linkNodes {
		for _, link := range append(node.GetParam("href"), node.GetParam("src")...) {
			fixedLink := fixLink(requestUrl.Scheme, requestUrl.Host, link)
			pos := strings.LastIndex(fixedLink, ".")
			if pos != -1 && webpageFileTypes.Contains(fixedLink[pos+1:]) {
				record.Files = append(record.Files, &rpb.File{FileUrl: fixedLink})
			} else {
				record.Links = append(record.Links, fixedLink)
			}
		}
	}
	w.logger.Debugf("Done loading page: %d byte(s), %d file link(s), %d other link(s)",
		len(body), len(record.Files), len(record.Links))
	return &rpb.RecordSet{Records: []*rpb.Record{record}}, nil
}

func NewWeb(name string, httpClient *http.Client) Chronicler {
	logger := util.NewLogger(name)
	if httpClient == nil {
		logger.Infof("No http client provided, using an own new one")
		httpClient = &http.Client{}
	}
	return &Web{
		name:   name,
		logger: logger,
		client: httpClient,
	}
}
