package frontend

import (
	"fmt"
	"io"
	"net/http"
	"testing"
	"time"
)

const (
	testingPort = 12345
)

func get(url string) ([]byte, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	data, err := io.ReadAll(resp.Body)
	return data, err
}

func TestFrontend(t *testing.T) {

	server := NewServer(testingPort, "static", nil)
	go (func() {
		if err := server.ListenAndServe(); err != nil {
			t.Fatalf("Could not start a server: %s", err)
		}
	})()

	time.Sleep(3 * time.Second)
	for _, tc := range []struct {
		desc string
		url  string
	}{
		{"simple test", fmt.Sprintf("http://localhost:%d/", testingPort)},
	} {
		t.Run(tc.desc, func(t *testing.T) {
			data, err := get(tc.url)
			if err != nil {
				t.Errorf("Could not fetch data from %s", tc.url)
			}
			if len(data) > 100 {
				data = data[:100]
			}
			fmt.Println(string(data))
		})
	}
}
