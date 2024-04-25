package resolver

import (
	"fmt"
	"io"
	"net/http"
	"net/url"

	cm "github.com/lanseg/golang-commons/common"

	"chronicler/status"
	"chronicler/storage"
)

const (
	userAgentCurl             = "curl/7.54"
	userAgentFirefox          = "Mozilla/5.0 (X11; Linux x86_64; rv:109.0) Gecko/20100101 Firefox/115.0"
	downloadQueueSize         = 10
	downloadBufferSize        = 1024 * 1024
	downloadProgressThreshold = 1024 * 1024 // Log the progress if content is bigger than that.
)

type httpDownloader struct {
	Downloader

	tasks      chan downloadTask
	stats      status.StatusClient
	storage    storage.Storage
	httpClient *http.Client
	logger     *cm.Logger
}

func (h *httpDownloader) copyData(src io.ReadCloser, dst io.WriteCloser, size int64) (int64, error) {
	defer src.Close()
	defer dst.Close()

	buffer := make([]byte, downloadBufferSize)
	totalWritten := int64(0)
	hasMore := true

	for hasMore {
		bytesRead, err := src.Read(buffer)
		if err == io.EOF {
			hasMore = false
		} else if err != nil {
			return totalWritten, err
		}

		bytesWritten, err := dst.Write(buffer[:bytesRead])
		if err != nil {
			return totalWritten, err
		}
		totalWritten += int64(bytesWritten)

		if size == -1 {
			h.logger.Debugf("Written %d of ??? bytes", totalWritten)
		} else {
			h.logger.Debugf("Written %d of %d bytes, %d remaining", totalWritten, size, size-totalWritten)
		}
	}
	return totalWritten, nil
}

func (h *httpDownloader) get(link string) (*http.Response, error) {
	req, err := http.NewRequest("GET", link, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", userAgentFirefox)
	return h.httpClient.Do(req)
}

func (h *httpDownloader) downloadLoop() {
	h.logger.Infof("Starting downloader loop")
	go (func() {
		for {
			task := <-h.tasks
			h.logger.Debugf("Started downloading %q to %q", task.source, task.id)
			u, err := url.Parse(task.source)
			if err != nil {
				h.logger.Warningf("Incorrect source url %s: %s", task.source, err)
				continue
			}

			metric := fmt.Sprintf("downloader.%s", cm.UUID4())
			h.stats.PutString(metric, u.String())
			resp, err := h.get(u.String())
			if err != nil {
				h.logger.Warningf("Cannot create get request for url %s: %s", u, err)
				h.stats.DeleteMetric(metric)
				continue
			}

			src := resp.Body
			h.logger.Debugf("Content size is: %d", resp.ContentLength)
			if err := h.storage.PutFile(task.id, task.source, src); err != nil {
				h.logger.Warningf("Error while writing data from %s to %s: %s", u, task.id, err)
			}
			h.stats.DeleteMetric(metric)
			src.Close()
		}
	})()
}

func (h *httpDownloader) ScheduleDownload(id string, source string) error {
	h.tasks <- downloadTask{id, source}
	h.logger.Infof("Scheduled new download %q to %q", source, id)
	return nil
}

func NewHttpDownloader(httpClient *http.Client, storage storage.Storage, stats status.StatusClient) Downloader {
	loader := &httpDownloader{
		storage:    storage,
		tasks:      make(chan downloadTask, downloadQueueSize),
		httpClient: httpClient,
		stats:      stats,
		logger:     cm.NewLogger("downloader"),
	}
	loader.downloadLoop()
	return loader
}
