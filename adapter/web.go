package adapter

import (
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strings"
	"time"

	rpb "chronicler/records/proto"
	"web/htmlparser"

	"github.com/lanseg/golang-commons/collections"
	cm "github.com/lanseg/golang-commons/common"
)

const (
	userAgent = "curl/7.54"
)

var (
	webpageFileTypes = collections.NewSet([]string{
		"jpg", "png", "js", "css", "json", "ico", "webp", "gif",
	})
)

type HttpClient interface {
	Do(request *http.Request) (*http.Response, error)
}

func isFileUrl(link string) bool {
	return webpageFileTypes.Contains(link)
}

func fixLink(scheme string, host string, link string) string {
	u, err := url.Parse(link)
	if err != nil {
		return link
	}
	if u.Host == "" {
		u.Host = host
	}
	if u.Scheme == "" {
		u.Scheme = scheme
	}
	return u.String()
}

func splitSrcset(srcset []string) []string {
	result := []string{}
	for _, src := range srcset {
		result = append(result, strings.Split(src, " ")[0])
	}
	return result
}

type webAdapter struct {
	timeSource func() time.Time
	logger     *cm.Logger
	client     HttpClient
}

func NewWebAdapter(httpClient HttpClient) Adapter {
	return createWebAdapter(httpClient, time.Now)
}

func createWebAdapter(httpClient HttpClient, timeSource func() time.Time) Adapter {
	logger := cm.NewLogger("WebAdapter")
	if httpClient == nil {
		logger.Infof("No http client provided, using an own new one")
		jar, err := cookiejar.New(nil)
		if err != nil {
			httpClient = &http.Client{}
			logger.Warningf("Got error while creating cookie jar %s", err.Error())
		} else {
			httpClient = &http.Client{
				Jar: jar,
			}
		}
	}
	return &webAdapter{
		logger:     logger,
		client:     httpClient,
		timeSource: timeSource,
	}
}

func (w *webAdapter) MatchLink(link string) *rpb.Source {
	if link == "" {
		return nil
	}
	u, err := url.Parse(link)
	if err != nil {
		return nil
	}
	return &rpb.Source{
		Url:  u.String(),
		Type: rpb.SourceType_WEB,
	}
}

func (w *webAdapter) Get(link string) (*http.Response, error) {
	req, err := http.NewRequest("GET", fixLink("https", "", link), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", userAgent)
	return w.client.Do(req)
}

func (w *webAdapter) SendMessage(*rpb.Message) {
	w.logger.Infof("Web adapter cannot send messages")
}

func (w *webAdapter) GetResponse(request *rpb.Request) []*rpb.Response {
	w.logger.Infof("Loading web page from %s", request.Target.Url)
	response, err := w.Get(request.Target.Url)
	if err != nil {
		return []*rpb.Response{{
			Request: request,
			Result:  []*rpb.RecordSet{},
		}}
	}
	defer response.Body.Close()

	requestUrl := response.Request.URL
	w.logger.Infof("Resolved actual URL as %s", requestUrl)
	body, err := io.ReadAll(response.Body)
	if err != nil {
		return []*rpb.Response{{
			Request: request,
			Result:  []*rpb.RecordSet{},
		}}
	}

	source := &rpb.Source{
		ChannelId: requestUrl.Host,
		Url:       requestUrl.String(),
		Type:      rpb.SourceType_WEB,
	}
	record := &rpb.Record{
		Source:      source,
		Time:        w.timeSource().Unix(),
		TextContent: htmlparser.StripTags(string(body)),
		RawContent:  body,
	}
	w.logger.Debugf("Parsing html content")
	root := htmlparser.ParseHtml(string(body))
	linkNodes := root.GetElementsByTagNames("a", "img", "script", "link", "source", "srcset")
	w.logger.Debugf("Found %d external link(s)", len(linkNodes))
	for _, node := range linkNodes {
		for _, link := range append(node.GetParam("href"), append(node.GetParam("src"), splitSrcset(node.GetParam("srcset"))...)...) {
			fixedLink := fixLink(requestUrl.Scheme, requestUrl.Host, link)
			pos := strings.LastIndex(fixedLink, ".")
			if pos != -1 && isFileUrl(fixedLink[pos+1:]) {
				record.Files = append(record.Files, &rpb.File{FileUrl: fixedLink})
			} else {
				record.Links = append(record.Links, fixedLink)
			}
		}
	}
	w.logger.Debugf("Done loading page: %d byte(s), %d file link(s), %d other link(s)",
		len(body), len(record.Files), len(record.Links))
	rs := &rpb.RecordSet{
		Id:           request.Id,
		Records:      []*rpb.Record{record},
		UserMetadata: []*rpb.UserMetadata{},
	}
	return []*rpb.Response{{
		Request: request,
		Result:  []*rpb.RecordSet{rs},
	}}
}
