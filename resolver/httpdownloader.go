package resolver

import (
	"bufio"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"

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

func (h *httpDownloader) Download(id string, source string) error {
	h.logger.Infof("Starting downloader loop")
	h.logger.Debugf("Started downloading %q to %q", source, id)
	u, err := url.Parse(source)
	if err != nil {
		return err
	}

	metric := fmt.Sprintf("downloader.%s", cm.UUID4())
	h.stats.PutString(metric, u.String())
	defer h.stats.DeleteMetric(metric)

	var src io.ReadCloser
	size := int64(0)
	if u.Scheme == "file" {
		u.Scheme = ""
		file, err := os.Open(u.String())
		if err != nil {
			return err
		}

		stat, err := file.Stat()
		if err != nil {
			return err
		}
		size = stat.Size()
		src = io.NopCloser(bufio.NewReader(file))
		defer src.Close()
	} else {
		resp, err := h.get(u.String())
		if err != nil {
			return err
		}
		src = resp.Body
		size = resp.ContentLength
		defer src.Close()
	}

	h.logger.Debugf("Content size is: %d", size)
	if err := h.storage.PutFile(id, source, src); err != nil {
		return err
	}
	return nil
}

func NewHttpDownloader(httpClient *http.Client, storage storage.Storage, stats status.StatusClient) Downloader {
	return &httpDownloader{
		storage:    storage,
		httpClient: httpClient,
		stats:      stats,
		logger:     cm.NewLogger("downloader"),
	}
}
