package downloader

import (
	"io"
	"net/http"
	"net/url"
	"os"

	cm "github.com/lanseg/golang-commons/common"
)

const (
	userAgentCurl             = "curl/7.54"
	userAgentFirefox          = "Mozilla/5.0 (X11; Linux x86_64; rv:109.0) Gecko/20100101 Firefox/115.0"
	downloadQueueSize         = 10
	downloadBufferSize        = 1024 * 1024
	downloadProgressThreshold = 1024 * 1024 // Log the progress if content is bigger than that.
)

type downloadTask struct {
	source string
	target string
}

type Downloader interface {
	ScheduleDownload(source string, target string) error
}

type httpDownloader struct {
	Downloader

	tasks      chan downloadTask
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
			h.logger.Debugf("Started downloading %q to %q", task.source, task.target)
			u, err := url.Parse(task.source)
			if err != nil {
				h.logger.Warningf("Incorrect source url %s: %s", task.source, err)
				continue
			}

			resp, err := h.get(u.String())
			if err != nil {
				h.logger.Warningf("Cannot create get request for url %s: %s", u, err)
				continue
			}

			dst, err := os.OpenFile(task.target, os.O_RDWR|os.O_CREATE, 0644)
			if err != nil {
				h.logger.Warningf("Cannot open write stream for local file %s: %s", task.target, err)
				dst.Close()
				continue
			}

			src := resp.Body
			h.logger.Debugf("Content size is: %d", resp.ContentLength)

			bytesWritten, err := h.copyData(src, dst, resp.ContentLength)
			if err != nil {
				h.logger.Warningf("Error while writing data from %s to %s: %s", u, task.target, err)
				src.Close()
			}
			h.logger.Debugf("Written %d byte(s) from %q to %q", bytesWritten, task.source, task.target)
		}
	})()
}

func (h *httpDownloader) ScheduleDownload(source string, target string) error {
	h.tasks <- downloadTask{source, target}
	h.logger.Infof("Scheduled new download %q to %q", source, target)
	return nil
}

func NewDownloader(httpClient *http.Client) Downloader {
	loader := &httpDownloader{
		tasks:      make(chan downloadTask, downloadQueueSize),
		httpClient: httpClient,
		logger:     cm.NewLogger("downloader"),
	}
	loader.downloadLoop()
	return loader
}

type noopDownloader struct {
	Downloader

	logger *cm.Logger
}

func (nd *noopDownloader) ScheduleDownload(source string, target string) error {
	nd.logger.Infof("Asked to download %s to %s, but doing nothing.", source, target)
	return nil
}

func NewNoopDownloader() Downloader {
	return &noopDownloader{
		logger: cm.NewLogger("noop-downloader"),
	}
}
