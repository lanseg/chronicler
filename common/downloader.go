package common

import (
	"io"
	"net/http"

	"chronicler/storage"
)

type Downloader interface {
	Download(string, storage.Storage) (int64, error)
}

type httpDownloader struct {
	Downloader

	client *http.Client
	logger Logger
}

func NewHttpDownloader(client *http.Client) Downloader {
	return &httpDownloader{
		logger: *NewLogger("HttpDownloader"),
		client: client,
	}
}

func (h *httpDownloader) Download(source string, target storage.Storage) (int64, error) {
	resp, err := h.client.Get(source)
	if err != nil {
		return -1, err
	}
	defer resp.Body.Close()

	targetFile, err := target.Put(&storage.PutRequest{Url: source})
	if err != nil {
		return -1, err
	}
	defer targetFile.Close()

	size, err := io.Copy(targetFile, resp.Body)
	if err != nil {
		return -1, err
	}
	return size, nil
}
