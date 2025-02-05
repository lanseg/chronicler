package adaptertest

import (
	"bytes"
	"chronicler/adapter"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/google/go-cmp/cmp"
	"google.golang.org/protobuf/testing/protocmp"

	opb "chronicler/proto"
)

type FakeHttpClient struct {
	adapter.HttpClient

	file string
}

func (fh *FakeHttpClient) Do(req *http.Request) (*http.Response, error) {
	bts, err := os.ReadFile(fh.file)
	if err != nil {
		return nil, err
	}
	return &http.Response{
		Body:    io.NopCloser(bytes.NewReader(bts)),
		Request: req,
	}, nil
}

func NewFakeHttp(responseFile string) *FakeHttpClient {
	return &FakeHttpClient{
		file: responseFile,
	}
}

func TestRequestResponse(a adapter.Adapter, link string, wantFile string) error {
	got, err := a.Get(&opb.Link{Href: link})
	if err != nil {
		return fmt.Errorf("error while doing get: %s", err)
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

	if diff := cmp.Diff(want, got, protocmp.Transform()); diff != "" {
		return fmt.Errorf("expected parsed and reference results to be equal, but got %s", diff)
	}
	return nil
}
