package common

import (
	"io"
	"net/http"
	"os"
)

type Downloader interface {
	Download(from string, to string) (int64, error)
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
func (h *httpDownloader) Download(sourcePath string, targetPath string) (int64, error) {
	h.logger.Debugf("Downloading %s to %s", sourcePath, targetPath)
	if _, err := os.Stat(targetPath); err == nil {
		h.logger.Debugf("File %s already exists as %s", sourcePath, targetPath)
		return -1, nil
	}

	targetFile, err := os.Create(targetPath)
	if err != nil {
		return -1, err
	}
	defer targetFile.Close()

	resp, err := h.client.Get(sourcePath)
	if err != nil {
		return -1, err
	}
	defer resp.Body.Close()

	size, err := io.Copy(targetFile, resp.Body)
	if err != nil {
		return -1, err
	}
	return size, nil
}
