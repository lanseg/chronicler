package common

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestHttpClient(t *testing.T) {
	svr := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(500 * time.Millisecond)
		fmt.Fprintf(w, "dummy data")
	}))
	defer svr.Close()

	t.Run("test download timeout", func(t *testing.T) {
		client := HttpClientBuilder{Timeout: 400 * time.Millisecond}.Build()
		req, err := http.NewRequest("GET", svr.URL, nil)
		if err != nil {
			t.Fatalf("Cannot create request: %q", err)
		}

		_, err = client.Do(req)
		if err == nil {
			t.Fatalf("Error while making request: %q", err)
		}
	})
}
