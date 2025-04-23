package common

import (
	"io"
	"net/http"
)

type Downloader interface {
	Download(source string, target io.Writer) (int64, error)
}

type httpDownloader struct {
	Downloader

	client HttpClient
}

func NewHttpDownloader(client HttpClient) Downloader {
	return &httpDownloader{
		client: client,
	}
}

func (h *httpDownloader) Download(source string, target io.Writer) (int64, error) {
	req, err := http.NewRequest("GET", source, nil)
	if err != nil {
		return 0, err
	}
	resp, err := h.client.Do(req)
	if err != nil {
		return -1, err
	}
	defer resp.Body.Close()

	size, err := io.Copy(target, resp.Body)
	if err != nil {
		return -1, err
	}
	return size, nil
}
