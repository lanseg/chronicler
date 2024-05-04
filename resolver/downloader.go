package resolver

import (
	cm "github.com/lanseg/golang-commons/common"
)

type downloadTask struct {
	id     string
	source string
}

type Downloader interface {
	Download(id string, source string) error
}

type noopDownloader struct {
	Downloader

	logger *cm.Logger
}

func (nd *noopDownloader) Download(source string, target string) error {
	nd.logger.Infof("Asked to download %s to %s, but doing nothing.", source, target)
	return nil
}

func NewNoopDownloader() Downloader {
	return &noopDownloader{
		logger: cm.NewLogger("noop-downloader"),
	}
}
