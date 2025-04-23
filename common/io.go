package common

import (
	"io"
	"mime"
	"net/url"
	"path/filepath"
)

type teeReadCloser struct {
	io.ReadCloser

	reader io.Reader
	closer io.Closer
}

func (rc *teeReadCloser) Read(p []byte) (n int, err error) {
	return rc.reader.Read(p)
}

func (rc *teeReadCloser) Close() error {
	return rc.closer.Close()
}

func GuessMimeType(href string) string {
	fileName := ""
	if u, err := url.Parse(href); err == nil {
		fileName = u.Path
	} else {
		fileName = href
	}
	return mime.TypeByExtension(filepath.Ext(fileName))
}
