package adapter

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"testing"
	"time"

	rpb "chronicler/records/proto"
)

const (
	webRequestUuid = "1a468cef-1368-408a-a20b-86b32d94a460"
)

type FakeHttpClient struct {
	HttpClient

	file string
}

func (fh *FakeHttpClient) Do(r *http.Request) (*http.Response, error) {
	bts, err := os.ReadFile(filepath.Join("testdata", fh.file))
	if err != nil {
		return nil, err
	}

	return &http.Response{
		Body:    io.NopCloser(bytes.NewReader(bts)),
		Request: r,
	}, nil
}

func NewFakeHttp(file string) HttpClient {
	return &FakeHttpClient{file: file}
}

func TestWebRequestResponse(t *testing.T) {
	for _, tc := range []struct {
		desc         string
		responseFile string
		resultFile   string
	}{
		{
			desc:         "Single update response",
			responseFile: "web_hello.html",
			resultFile:   "web_hello_record.json",
		},
	} {
		t.Run(tc.desc, func(t *testing.T) {
			web := createWebAdapter(NewFakeHttp(tc.responseFile),
				func() time.Time {
					return time.Date(2009, time.November, 10, 23, 0, 0, 0, time.UTC)
				})
			web.SubmitRequest(&rpb.Request{
				Id:     webRequestUuid,
				Target: &rpb.Source{Url: "google.com"},
			})
			ups := web.GetRecordSet()

			want := &rpb.RecordSet{}
			if err := readJson(tc.resultFile, want); err != nil {
				t.Errorf("Cannot load json with an expected result \"%s\": %s", tc.resultFile, err)
			}
			if fmt.Sprintf("%+v", want) != fmt.Sprintf("%+v", ups) {
				t.Errorf("Expected result to be:\n%+v\nBut got:\n%+v", want, ups)
			}
		})
	}
}
