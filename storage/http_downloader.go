package storage

import (
	"io"
	"net/http"
	"net/url"

    cm "github.com/lanseg/golang-commons/common" 
)

const (
	userAgent = "curl/7.54"
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
	req.Header.Set("User-Agent", userAgent)
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
		client = &http.Client{}
	}
	return &HttpDownloader{
		httpClient: client,
		logger:     logger,
	}
}
