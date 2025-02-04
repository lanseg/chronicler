package fourchan

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

	"github.com/google/go-cmp/cmp"
	"google.golang.org/protobuf/testing/protocmp"
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

func TestFourchanAdapter(t *testing.T) {
	for _, tc := range []struct {
		name string
		file string
	}{
		{name: "simple post", file: "g104208157"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			// What we get
			fakeClient := newFakeHttp(tc.file + "_chan.json")
			fourchanAdapter := NewAdapter(fakeClient)
			got, err := fourchanAdapter.Get(&opb.Link{Href: "https://boards.4chan.org/g/thread/104191633"})
			if err != nil {
				t.Errorf("error while doing get: %s", err)
				return
			}
			gotBytes, _ := json.Marshal(got)
			os.WriteFile(fmt.Sprintf("/tmp/%s.json", tc.file), gotBytes, 0777)

			// Reference data
			wantBytes, err := os.ReadFile(fmt.Sprintf("test_data/%s.json", tc.file))
			if err != nil {
				t.Errorf("cannot load reference file %s.json: %s", tc.file, err)
				return
			}
			want := []*opb.Object{}
			if err = json.Unmarshal(wantBytes, &want); err != nil {
				t.Errorf("cannot unmarshal reference file %s.json: %s", tc.file, err)
				return
			}

			if diff := cmp.Diff(want, got, protocmp.Transform()); diff != "" {
				t.Errorf("expected parsed and reference results to be equal, but got %s", diff)
			}
		})
	}
}

func TestFourChanAdapterMatcher(t *testing.T) {
	fourchanAdapter := NewAdapter(nil)
	for _, tc := range []struct {
		name    string
		url     string
		matches bool
	}{
		{name: "basic http link", url: `https://boards.4chan.org/g/thread/102519935/this-board-is-for-the-discussion-of-technology`, matches: true},
		{name: "link with id", url: `https://boards.4chan.org/g/thread/104191633#p104206868`, matches: true},
		{name: "link with name", url: `https://boards.4chan.org/g/thread/104205910/termux`, matches: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if tc.matches != fourchanAdapter.Match(&opb.Link{Href: tc.url}) {
				t.Errorf("web adapter should match %s", tc.url)
			}
		})
	}
}
