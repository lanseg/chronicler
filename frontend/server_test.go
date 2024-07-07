package frontend

import (
	"fmt"
	"io"
	"net/http"
	"testing"
	"time"

	opt "github.com/lanseg/golang-commons/optional"

	rpb "chronicler/records/proto"
	"chronicler/status"
	"chronicler/storage"
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

type fakeStorage struct {
	storage.NoOpStorage
}

func (fs *fakeStorage) GetRecordSet(id string) opt.Optional[*rpb.RecordSet] {
	return opt.Nothing[*rpb.RecordSet]{}
}

func TestFrontend(t *testing.T) {
	fakeStorage := &fakeStorage{}
	stats, _ := status.NewNoopStatusClient("")
	server := NewServer(testingPort, "static", fakeStorage, stats)
	go (func() {
		if err := server.ListenAndServe(); err != nil {
			t.Fatalf("Could not start a server: %s", err)
		}
	})()

	time.Sleep(3 * time.Second)
	for _, tc := range []struct {
		desc string
		url  string
		want string
	}{
		{
			desc: "simple test",
			url:  fmt.Sprintf("http://localhost:%d/", testingPort),
		},
		{
			desc: "Error for an empty record",
			url:  fmt.Sprintf("http://localhost:%d/chronicler/records/%s", testingPort, "somerecord"),
			want: "No record with id \"somerecord\"\n",
		},
	} {
		t.Run(tc.desc, func(t *testing.T) {
			data, err := get(tc.url)
			if err != nil {
				t.Errorf("Could not fetch data from %s", tc.url)
			}
			if len(data) > 100 {
				data = data[:100]
			}

			doc := string(data)
			if tc.want != "" && tc.want != doc {
				t.Errorf("expected %q, but got %q", tc.want, doc)
			}
		})
	}
}
