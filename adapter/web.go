package adapter

import (
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strings"
	"time"

	"chronicler/records"
	rpb "chronicler/records/proto"
	"chronicler/webdriver"

	"github.com/lanseg/golang-commons/almosthtml"
	"github.com/lanseg/golang-commons/collections"
	cm "github.com/lanseg/golang-commons/common"
)

const (
	userAgent = "curl/7.54"
)

var (
	webpageFileTypes = collections.NewSet([]string{
		// images
		"jpg", "png", "ico", "webp", "gif",
		// page content
		"js", "css", "json",
		// video
		"mp4", "webm",
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

func splitSrcset(srcset string) []string {
	return []string{strings.Split(srcset, " ")[0]}
}

type webAdapter struct {
	timeSource func() time.Time
	logger     *cm.Logger
	client     HttpClient
	browser    webdriver.WebdriverService
}

func NewWebAdapter(httpClient HttpClient, browser webdriver.WebdriverService) Adapter {
	return createWebAdapter(httpClient, browser, time.Now)
}

func createWebAdapter(httpClient HttpClient, browser webdriver.WebdriverService, timeSource func() time.Time) Adapter {
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
		browser:    browser,
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

func (w *webAdapter) SendMessage(*rpb.Message) {
	w.logger.Infof("Web adapter cannot send messages")
}

func (w *webAdapter) GetResponse(request *rpb.Request) []*rpb.Response {
	w.logger.Infof("Loading web page from %s", request.Target.Url)

	body := []byte{}
	actualUrl := request.Target.Url
	w.browser.RunSession(func(d webdriver.WebDriver) {
		d.Navigate(request.Target.Url)
		d.GetCurrentURL().IfPresent(func(url string) {
			actualUrl = url
		})
		d.GetPageSource().IfPresent(func(bodyStr string) {
			body = []byte(bodyStr)
		})
	})

	requestUrl, _ := url.Parse(fixLink("https", "", actualUrl))
	w.logger.Infof("Resolved actual URL as %s", requestUrl)
	record := records.NewRecord(&rpb.Record{
		Source: &rpb.Source{
			ChannelId: requestUrl.Host,
			Url:       requestUrl.String(),
			Type:      rpb.SourceType_WEB,
		},
		Time:        w.timeSource().Unix(),
		TextContent: almosthtml.StripTags(string(body)),
		RawContent:  body,
	})

	w.logger.Debugf("Parsing html content")
	root, _ := almosthtml.ParseHTML(string(body))
	knownLinks := collections.NewSet([]string{})
	linkNodes := root.GetElementsByTags("a", "img", "script", "link", "source", "srcset")

	for _, node := range linkNodes {
		for _, link := range append(splitSrcset(node.Params["srcset"]), node.Params["href"], node.Params["src"]) {
			fixedLink := fixLink(requestUrl.Scheme, requestUrl.Host, link)
			if knownLinks.Contains(fixedLink) {
				continue
			}
			knownLinks.Add(fixedLink)
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
	rs := records.NewRecordSet(&rpb.RecordSet{
		Id:           cm.UUID4(),
		Records:      []*rpb.Record{record},
		UserMetadata: []*rpb.UserMetadata{},
	})
	return []*rpb.Response{{
		Request: request,
		Result:  []*rpb.RecordSet{rs},
	}}
}
