package pikabu

import (
	"bytes"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"testing"
)

type fakeHttpClient struct {
	data []byte
}

func (c *fakeHttpClient) Get(string) (resp *http.Response, err error) {
	return &http.Response{
		Body: io.NopCloser(bytes.NewReader(c.data)),
	}, nil
}

func TestPikabuSourceProvider(t *testing.T) {

	t.Run("Creates", func(t *testing.T) {
		disputed, err := os.ReadFile(filepath.Join("testdata", "pikabu_disputed.html"))
		if err != nil {
			t.Errorf("Cannot open sample file: %s", err)
			return
		}

		srcs := NewDisputedProvider(&fakeHttpClient{
			data: disputed,
		}).GetSources()
		if len(srcs) != 11 {
			t.Errorf("Expected to return 12 sources, but got %d", len(srcs))
		}
	})
}
