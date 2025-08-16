package common

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/cookiejar"
	"os"
	"path/filepath"
)

type HttpClient interface {
	Do(request *http.Request) (*http.Response, error)
}

type HttpClientBuilder struct {
	CookieJarPath string
	CacheDirPath  string
}

func (hbc HttpClientBuilder) Build() HttpClient {
	baseClient := &http.Client{}
	if hbc.CookieJarPath != "" {
		jar, err := cookiejar.New(&cookiejar.Options{})
		if err != nil {
			log.Fatal(err)
		}
		baseClient.Jar = jar
	}

	logger := NewLogger("HttpClient")
	client := HttpClient(baseClient)
	if hbc.CacheDirPath == "" {
		logger.Infof("Using http client without cache")
	} else {
		logger.Infof("Using http client cache at %q", hbc.CacheDirPath)
		client = &cachingHttpClient{
			parent:    client,
			logger:    logger,
			cachePath: hbc.CacheDirPath,
		}
	}

	return client
}

func makeCachedResponse(localPath string, request *http.Request) (*http.Response, error) {
	status := 200
	body, err := os.ReadFile(localPath)
	if err != nil {
		status = 500
		body = []byte(err.Error())
	}
	return &http.Response{
		Status:        fmt.Sprintf("%d", status),
		StatusCode:    status,
		Body:          io.NopCloser(bytes.NewReader(body)),
		Request:       request,
		ContentLength: int64(len(body)),
	}, nil
}

type cachingHttpClient struct {
	HttpClient

	logger    *Logger
	cachePath string
	parent    HttpClient
}

func (cc *cachingHttpClient) Do(request *http.Request) (*http.Response, error) {
	cacheRoot := filepath.Join(cc.cachePath, "chronicler")
	if err := os.MkdirAll(cacheRoot, 0766); err != nil {
		return nil, err
	}

	cachedPath := filepath.Join(cacheRoot, Sanitize(request.URL.String(), 255))
	if _, err := os.Stat(cachedPath); err == nil {
		cc.logger.Debugf("loaded %q from cache at %q", request.URL.String(), cachedPath)
		return makeCachedResponse(cachedPath, request)
	}

	response, err := cc.parent.Do(request)
	if err != nil {
		return nil, err
	}
	if response.StatusCode != 200 {
		return response, nil
	}
	if target, err := os.Create(cachedPath); err == nil {
		cc.logger.Debugf("saving %q to cache at %q", request.URL.String(), cachedPath)
		response.Body = &teeReadCloser{
			reader: io.TeeReader(response.Body, target),
			closer: response.Body,
		}
	}
	return response, nil
}
