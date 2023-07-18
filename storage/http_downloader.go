package storage

import (
	"io"
	"net/http"
	"net/url"
	"os"

	"chronicler/util"
)

const (
	userAgent = "curl/7.54"
)

type HttpDownloader struct {
	httpClient *http.Client

	logger *util.Logger
}

func copyReader(src io.ReadCloser, dst string) error {
	defer src.Close()

	targetFile, err := os.Create(dst)
	defer targetFile.Close()
	if err != nil {
		return err
	}
	_, err = io.Copy(targetFile, src)
	if err != nil {
		return err
	}
	return nil
}

func (d *HttpDownloader) get(link string) (*http.Response, error) {
	req, err := http.NewRequest("GET", link, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", userAgent)
	return d.httpClient.Do(req)
}

func (d *HttpDownloader) Download(link string, target string) error {
	d.logger.Debugf("Downloading file from %s to %s", link, target)
	u, err := url.Parse(link)
	if err != nil {
		return err
	}
	if u.Scheme == "" {
		u.Scheme = "https"
	}
	resp, err := d.get(u.String())
	if err != nil {
		return err
	}
	return copyReader(resp.Body, target)
}

func NewHttpDownloader(client *http.Client) *HttpDownloader {
	logger := util.NewLogger("storage")
	if client == nil {
		logger.Debugf("No http client provided for Downloader. Using custom one.")
		client = &http.Client{}
	}
	return &HttpDownloader{
		httpClient: client,
		logger:     logger,
	}
}
