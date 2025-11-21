package common

import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestDownloader(t *testing.T) {

	t.Run("test download with timeout", func(t *testing.T) {
		svr := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			time.Sleep(1000 * time.Millisecond)
			fmt.Fprintf(w, "dummy data")
		}))

		buf := bytes.Buffer{}
		loader := NewHttpDownloader(HttpClientBuilder{Timeout: 400 * time.Millisecond}.Build())
		_, err := loader.Download(svr.URL, &buf)
		if err == nil {
			t.Fatalf("Expected timeout error, got nil")
		}
		defer svr.Close()
	})
}
