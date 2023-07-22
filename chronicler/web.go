package chronicler

import (
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strings"
	"time"

	rpb "chronicler/proto/records"
	"chronicler/util"
	"web/htmlparser"

	"github.com/lanseg/golang-commons/collections"
)

const (
	userAgent = "curl/7.54"
)

var (
	webpageFileTypes = collections.NewSet([]string{
		"jpg", "png", "js", "css", "json", "ico", "webp", "gif",
	})
)

func isFileUrl(link string) bool {
	if webpageFileTypes.Contains(link) {
		return true
	}
	u, err := url.Parse(link)
	if err != nil {
		return false
	}
	return u.Scheme == "data"
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

type Web struct {
	Chronicler

	requests chan *rpb.Request
	records  chan *rpb.RecordSet
	logger   *util.Logger
	client   *http.Client
}

func (w *Web) GetRecordSource() *rpb.RecordSet {
	return <-w.records
}

func (w *Web) SubmitRequest(r *rpb.Request) {
	w.requests <- r
}

func (w *Web) Get(link string) (*http.Response, error) {
	req, err := http.NewRequest("GET", fixLink("https", "", link), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", userAgent)
	return w.client.Do(req)
}

func (w *Web) requestLoop() {
	w.logger.Infof("Starting request loop")
	for {
		request := <-w.requests
		w.logger.Infof("New request: %s", request)
		r, err := w.getRecords(request)
		if err != nil {
			w.logger.Warningf("Error while doing a request: %s", err)
		} else {
			w.records <- r
		}
	}
}

func (w *Web) getRecords(request *rpb.Request) (*rpb.RecordSet, error) {
	w.logger.Infof("Loading web page from %s", request.Source.Url)
	response, err := w.Get(request.Source.Url)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	requestUrl := response.Request.URL
	w.logger.Infof("Resolved actual URL as %s", requestUrl)
	body, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}

	source := &rpb.Source{
		ChannelId: requestUrl.Host,
		Url:       requestUrl.String(),
		Type:      rpb.SourceType_WEB,
	}
	record := &rpb.Record{
		Source:      source,
		Time:        time.Now().Unix(),
		TextContent: string(body),
	}
	w.logger.Debugf("Parsing html content")
	root := htmlparser.ParseHtml(record.TextContent)
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
	return &rpb.RecordSet{
		Request: &rpb.Request{
			Source: source,
			Parent: request.Parent,
		},
		Records: []*rpb.Record{record},
	}, nil
}

func NewWeb(httpClient *http.Client) *Web {
	logger := util.NewLogger("Web")
	if httpClient == nil {
		logger.Infof("No http client provided, using an own new one")
		httpClient = &http.Client{}

		jar, err := cookiejar.New(nil)
		if err != nil {
			logger.Warningf("Got error while creating cookie jar %s", err.Error())
		} else {
			httpClient.Jar = jar
		}
	}
	result := &Web{
		logger:   logger,
		client:   httpClient,
		records:  make(chan *rpb.RecordSet),
		requests: make(chan *rpb.Request),
	}
	go result.requestLoop()
	return result
}
