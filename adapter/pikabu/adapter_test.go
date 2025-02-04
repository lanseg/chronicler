package pikabu

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"google.golang.org/protobuf/testing/protocmp"

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

func TestPikabuAdapter(t *testing.T) {
	for _, tc := range []struct {
		name string
		file string
	}{
		{name: "simple post", file: "12333399"},
		{name: "post with comments", file: "12335516"},
		{name: "post with comments", file: "12336257"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			// What we get
			fakeClient := newFakeHttp(tc.file + ".html")
			pikabuAdapter := NewAdapter(fakeClient)
			got, err := pikabuAdapter.Get(&opb.Link{Href: fmt.Sprintf("http://pikabu.ru/story/_%s", tc.file)})
			if err != nil {
				t.Errorf("error while doing get: %s", err)
				return
			}

			data, _ := json.Marshal(got)
			os.WriteFile(fmt.Sprintf("/tmp/%s.json", tc.file), data, 0777)
			// Reference data
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

			if diff := cmp.Diff(want, got, protocmp.Transform(),
				cmpopts.SortSlices(func(a, b string) bool { return a < b })); diff != "" {
				t.Errorf("expected parsed and reference results to be equal, but got %s", diff)
			}
		})
	}
}

func TestWebAdapterMatcher(t *testing.T) {
	pikabuAdapter := NewAdapter(nil)
	for _, tc := range []struct {
		name    string
		url     string
		matches bool
	}{
		{name: "basic http link", url: `http://localhost.localdomain`, matches: false},
		{name: "basic broken http link", url: `ht#t\];]ldomain`, matches: false},
		{name: "empty link", url: ``, matches: false},
		{name: "link without schema", url: `localhost.localdomain`, matches: false},
		{name: "link without all parameters", url: `localhost.localdomain`, matches: false},
		{name: "link without post id", url: "https://pikabu.ru/story/", matches: false},
		{name: "link with post id", url: "https://pikabu.ru/story/_1234", matches: true},
		{name: "link with post name and id", url: "https://pikabu.ru/story/a_post_name_and_12333062", matches: true},
		{name: "link with post id and query args", url: "https://pikabu.ru/story/bwef_wef_12333062?cid=339516221", matches: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if tc.matches != pikabuAdapter.Match(&opb.Link{Href: tc.url}) {
				t.Errorf("web adapter should match %s", tc.url)
			}
		})
	}
}
