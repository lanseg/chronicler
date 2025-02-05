package fourchan

import (
	"path/filepath"
	"testing"

	"chronicler/adapter/adaptertest"
	opb "chronicler/proto"
)

func TestFourchanAdapter(t *testing.T) {
	for _, tc := range []struct {
		name string
		file string
	}{
		{name: "simple post", file: "g104208157"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if err := adaptertest.TestRequestResponse(
				NewAdapter(adaptertest.NewFakeHttp(filepath.Join("test_data", tc.file+"_chan.json"))),
				"https://boards.4chan.org/g/thread/104191633",
				filepath.Join("test_data", tc.file+".json")); err != nil {
				t.Error(err)
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
