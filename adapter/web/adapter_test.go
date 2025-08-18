package web

import (
	"path/filepath"
	"testing"

	"chronicler/adapter/adaptertest"
	opb "chronicler/proto"
)

func TestWebAdapterMatcher(t *testing.T) {
	webAdapter := NewAdapter(nil, &WebSettings{true})
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
		name      string
		file      string
		recursive bool
	}{
		{name: "simple html", file: "simple.html", recursive: true},
		{name: "self link html", file: "self_link.html", recursive: true},
		{name: "broken links", file: "broken_links.html", recursive: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			response := filepath.Join("test_data", tc.file+".json")
			http := adaptertest.NewFakeHttp(filepath.Join("test_data", tc.file))
			adapter := NewAdapter(http, &WebSettings{tc.recursive})
			if err := adaptertest.TestRequestResponse(adapter, "http://demo", response); err != nil {
				t.Error(err)
			}
		})
	}
}
