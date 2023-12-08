package storage

import (
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/url"

	cm "github.com/lanseg/golang-commons/common"
)

const (
	userAgentCurl    = "curl/7.54"
	userAgentFirefox = "Mozilla/5.0 (X11; Linux x86_64; rv:109.0) Gecko/20100101 Firefox/115.0"
)

type HttpDownloader struct {
	httpClient *http.Client

	logger *cm.Logger
}

func (d *HttpDownloader) get(link string) (*http.Response, error) {
	req, err := http.NewRequest("GET", link, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", userAgentFirefox)
	return d.httpClient.Do(req)
}

func (d *HttpDownloader) Download(link string) ([]byte, error) {
	d.logger.Debugf("Downloading file from %s", link)
	u, err := url.Parse(link)
	if err != nil {
		return nil, err
	}
	if u.Scheme == "" {
		u.Scheme = "https"
	}
	resp, err := d.get(u.String())
	if err != nil {
		return nil, err
	}
	return io.ReadAll(resp.Body)
}

func NewHttpDownloader(client *http.Client) *HttpDownloader {
	logger := cm.NewLogger("storage")
	if client == nil {
		logger.Debugf("No http client provided for Downloader. Using custom one.")
		jar, err := cookiejar.New(nil)
		if err != nil {
			logger.Debugf("Using http client without cookie jar because of error: %s", err)
			client = &http.Client{}
		} else {
			client = &http.Client{
				Jar: jar,
			}
		}
	}
	return &HttpDownloader{
		httpClient: client,
		logger:     logger,
	}
}
