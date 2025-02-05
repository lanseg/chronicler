package pikabu

import (
	"fmt"
	"path/filepath"
	"testing"

	"chronicler/adapter/adaptertest"
	opb "chronicler/proto"
)

func TestPikabuAdapter(t *testing.T) {
	for _, tc := range []struct {
		name string
		file string
	}{
		{name: "simple post", file: "12333399"},
		{name: "post with comments", file: "12335516"},
		{name: "post with comments and more content", file: "12335104"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			// What we get
			if err := adaptertest.TestRequestResponse(
				NewAdapter(adaptertest.NewFakeHttp(filepath.Join("test_data", tc.file+".html"))),
				fmt.Sprintf("http://pikabu.ru/story/_%s", tc.file),
				filepath.Join("test_data", tc.file+".json")); err != nil {
				t.Error(err)
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
