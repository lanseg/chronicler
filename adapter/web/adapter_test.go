package web

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"testing"

	opb "chronicler/proto"
)

type fakeHttpClient struct {
	HttpClient

	file string
}

func (fh *fakeHttpClient) Do(req *http.Request) (*http.Response, error) {
	bts, err := os.ReadFile(filepath.Join("test_data", fh.file))
	if err != nil {
		return nil, err
	}
	return &http.Response{
		Body:    io.NopCloser(bytes.NewReader(bts)),
		Request: req,
	}, nil
}

func newFakeHttp(file string) *fakeHttpClient {
	return &fakeHttpClient{
		file: file,
	}
}

func TestWebAdapterMatcher(t *testing.T) {
	webAdapter := NewAdapter(nil)
	for _, tc := range []struct {
		name    string
		url     string
		matches bool
	}{
		{name: "basic http link", url: `http://localhost.localdomain`, matches: true},
		{name: "basic broken http link", url: `htt\];]'/localhost.localdomain`, matches: false},
		{name: "link without schema", url: `localhost.localdomain`, matches: false},
		{name: "link without all parameters", url: `localhost.localdomain`, matches: false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if tc.matches != webAdapter.Match(&opb.Link{Href: tc.url}) {
				t.Errorf("web adapter should match %s", tc.url)
			}
		})
	}
}

func TestWebAdapter(t *testing.T) {
	for _, tc := range []struct {
		name string
		file string
	}{
		{name: "simple html", file: "simple.html"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			want := []*opb.Object{}
			wantBytes, err := os.ReadFile(fmt.Sprintf("test_data/%s.json", tc.file))
			if err != nil {
				t.Errorf("cannot load reference file %s.json: %s", tc.file, err)
				return
			}
			if err = json.Unmarshal(wantBytes, &want); err != nil {
				t.Errorf("cannot unmarshal reference file %s.json: %s", tc.file, err)
				return
			}

			fakeClient := newFakeHttp(tc.file)
			webAdapter := NewAdapter(fakeClient)
			got, err := webAdapter.Get(&opb.Link{Href: "http://demo"})
			if err != nil {
				t.Errorf("error while doing get: %s", err)
				return
			}
			if fmt.Sprintf("%s", want) != fmt.Sprintf("%v", got) {
				t.Errorf("expected result to be %q, but got %q", want, got)
			}
		})
	}
}
