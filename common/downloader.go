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

	client *http.Client
}

func NewHttpDownloader(client *http.Client) Downloader {
	return &httpDownloader{
		client: client,
	}
}

func (h *httpDownloader) Download(source string, target io.Writer) (int64, error) {
	resp, err := h.client.Get(source)
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
