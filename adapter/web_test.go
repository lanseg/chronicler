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
	"chronicler/webdriver"
)

const (
	webRequestUuid = "1a468cef-1368-408a-a20b-86b32d94a460"
)

var (
	fakeTime = time.Date(2009, time.November, 10, 23, 0, 0, 0, time.UTC)
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

func newFakeHttp(file string) HttpClient {
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
			web := createWebAdapter(
				newFakeHttp(tc.responseFile),
				webdriver.WrapExclusive(newFakeWebdriver(tc.responseFile)),
				func() time.Time { return fakeTime })

			ups := web.GetResponse(&rpb.Request{
				Id:     webRequestUuid,
				Target: &rpb.Source{Url: "google.com"},
			})[0].Result[0]

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

func newWebSrc(url string) *rpb.Source {
	return &rpb.Source{
		Url:  url,
		Type: rpb.SourceType_WEB,
	}
}

func TestWebLinkMatcher(t *testing.T) {
	for _, tc := range []struct {
		desc string
		link string
		want *rpb.Source
	}{
		{
			desc: "link with and postfix prefix matches",
			link: "http://somelink.com/whatever?param&b=c",
			want: newWebSrc("http://somelink.com/whatever?param&b=c"),
		},
		{
			desc: "double slash link matches",
			link: "//somelink.com",
			want: newWebSrc("//somelink.com"),
		},
		{
			desc: "empty string doesnt match",
			link: "",
			want: nil,
		},
	} {
		t.Run(tc.desc, func(t *testing.T) {
			tg := NewWebAdapter(nil, webdriver.WrapExclusive(&webdriver.NoopWebdriver{}))

			result := tg.MatchLink(tc.link)
			if fmt.Sprintf("%+v", tc.want) != fmt.Sprintf("%+v", result) {
				t.Errorf("Expected result to be:\n%+v\nBut got:\n%+v", tc.want, result)
			}
		})
	}
}
