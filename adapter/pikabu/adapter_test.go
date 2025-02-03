package pikabu

import (
	"testing"

	opb "chronicler/proto"
)

func TestWebAdapterMatcher(t *testing.T) {
	webAdapter := NewAdapter(nil)
	for _, tc := range []struct {
		name    string
		url     string
		matches bool
	}{
		{name: "basic http link", url: `http://localhost.localdomain`, matches: false},
		{name: "basic broken http link", url: `htt\];]'/localhost.localdomain`, matches: false},
		{name: "link without schema", url: `localhost.localdomain`, matches: false},
		{name: "link without all parameters", url: `localhost.localdomain`, matches: false},
		{name: "link without post id", url: "https://pikabu.ru/story/", matches: false},
		{name: "link with post id", url: "https://pikabu.ru/story/_1234", matches: true},
		{name: "link with post name and id", url: "https://pikabu.ru/story/a_post_name_and_12333062", matches: true},
		{name: "link with post id and query args", url: "https://pikabu.ru/story/bwef_wef_12333062?cid=339516221", matches: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if tc.matches != webAdapter.Match(&opb.Link{Href: tc.url}) {
				t.Errorf("web adapter should match %s", tc.url)
			}
		})
	}
}
