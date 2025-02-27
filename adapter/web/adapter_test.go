package web

import (
	"path/filepath"
	"testing"

	"chronicler/adapter/adaptertest"
	opb "chronicler/proto"
)

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
		//{name: "simple html", file: "simple.html"},
		//{name: "self link html", file: "self_link.html"},
		{name: "created at parse", file: "broken_links.html"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if err := adaptertest.TestRequestResponse(
				NewAdapter(adaptertest.NewFakeHttp(filepath.Join("test_data", tc.file))),
				"http://demo",
				filepath.Join("test_data", tc.file+".json")); err != nil {
				t.Error(err)
			}
		})
	}
}
