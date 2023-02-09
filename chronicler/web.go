package chronicler

import (
	rpb "chronicler/proto/records"
	"chronicler/util"
	"io"
	"net/http"
	"net/url"
	"strings"
	"unicode"
)

type Web struct {
	Chronicler

	name   string
	logger *util.Logger
	client *http.Client
}

func (t *Web) GetName() string {
	return t.name
}

func (w *Web) GetRecords(request *rpb.Request) (*rpb.RecordSet, error) {
	response, err := http.Get(request.Source.Url)
	if err != nil {
		return nil, err
	}
	requestUrl := response.Request.URL
	defer response.Body.Close()

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}
	html := string(body)
	links := w.parseHtml(html)
	linkedFiles := []*rpb.File{}
	for _, link := range links {
		linkUrl, err := url.Parse(link)
		if err != nil {
			w.logger.Warningf("Looks like a malformed url: %s", linkUrl.String())
		}
		if linkUrl.Scheme == "" {
			linkUrl.Scheme = requestUrl.Scheme
		}
		if linkUrl.Host == "" {
			linkUrl.Host = requestUrl.Host
		}
		if linkUrl.Scheme == "mailto" || linkUrl.Scheme == "tel" {
			continue
		}
		linkedFiles = append(linkedFiles, &rpb.File{
			FileUrl: linkUrl.String(),
		})
	}
	return &rpb.RecordSet{
		Records: []*rpb.Record{
			{
				Source: &rpb.Source{
					ChannelId: requestUrl.Host,
				},
				Files:       linkedFiles,
				TextContent: html,
			},
		},
	}, nil
}

func (w *Web) parseHtml(html string) []string {
	links := []string{}
	tokens := []string{}
	buffer := ""
	inTag := false
	inQuote := false
	for _, ch := range html {
		if ch == '<' {
			inTag = true
			buffer = ""
		}
		if !inTag {
			continue
		}
		if ch == '\\' {
			buffer += string(ch)
			continue
		}
		if ch == '"' {
			inQuote = !inQuote
		}
		if inQuote {
			buffer += string(ch)
			continue
		}
		if (ch == '=' || ch == '>' || unicode.IsSpace(ch)) && buffer != "" {
			tokens = append(tokens, buffer)
			if ch == '=' {
				tokens = append(tokens, "=")
			}
			buffer = ""
		} else if ch != '<' && !unicode.IsSpace(ch) {
			buffer += string(ch)
		}
		if ch == '>' {
			for i, value := range tokens {
				if i < len(tokens)-2 && tokens[i+1] == "=" && (value == "src" || value == "href") {
					link := strings.Trim(tokens[i+2], "\"")
					if link != "#" {
						links = append(links, link)
					}
				}
			}
			buffer = ""
			inTag = false
			tokens = []string{}
		}
	}
	return links
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
