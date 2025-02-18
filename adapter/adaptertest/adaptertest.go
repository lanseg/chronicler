package adaptertest

import (
	"bytes"
	"chronicler/adapter"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"google.golang.org/protobuf/testing/protocmp"

	opb "chronicler/proto"
)

type FakeHttpClient struct {
	adapter.HttpClient

	currnetFile int
	files       []string
}

func (fh *FakeHttpClient) Do(req *http.Request) (*http.Response, error) {
	bts, err := os.ReadFile(fh.files[fh.currnetFile])
	fh.currnetFile = (fh.currnetFile + 1) % len(fh.files)
	if err != nil {
		return nil, err
	}
	return &http.Response{
		Body:    io.NopCloser(bytes.NewReader(bts)),
		Request: req,
	}, nil
}

func NewFakeHttp(responseFile ...string) *FakeHttpClient {
	return &FakeHttpClient{
		files: responseFile,
	}
}

func TestRequestResponse(a adapter.Adapter, link string, wantFile string) error {
	got, err := a.Get(&opb.Link{Href: link})
	if err != nil {
		return fmt.Errorf("error while doing get: %s", err)
	}
	if strings.Contains(link, "reddit") {
		bu, _ := json.Marshal(got)
		os.WriteFile("/home/arusakov/devel/lanseg/chronicler/reddit_test_out.json", bu, 0777)
	}
	// Reference data
	wantBytes, err := os.ReadFile(wantFile)
	if err != nil {
		return fmt.Errorf("cannot load reference file %s: %s", wantFile, err)
	}
	want := []*opb.Object{}
	if err = json.Unmarshal(wantBytes, &want); err != nil {
		return fmt.Errorf("cannot unmarshal reference file %s: %s", wantFile, err)

	}

	options := []cmp.Option{
		protocmp.Transform(),
		cmpopts.SortSlices(func(a, b string) bool { return a < b }),
	}
	if diff := cmp.Diff(want, got, options...); diff != "" {
		return fmt.Errorf("expected parsed and reference results to be equal, but got %s", diff)
	}
	return nil
}
