package reddit

import (
	"chronicler/adapter/adaptertest"
	opb "chronicler/proto"
	"path/filepath"
	"testing"
)

func TestRedditAdapterMatcher(t *testing.T) {
	redditAdapter := NewAdapter(nil)
	for _, tc := range []struct {
		name    string
		url     string
		matches bool
	}{
		{
			name:    "successful match",
			url:     "https://www.reddit.com/r/law/comments/1inaszr/musk_crashes_trumps_interview_and_goes_on_an_info/",
			matches: true,
		},
		{
			name:    "comment succesful match",
			url:     "https://www.reddit.com/r/law/comments/1inaszr/comment/mc9uo5g/?utm_source=share&utm_medium=web3x&utm_name=web3xcss&utm_term=1&utm_content=share_button",
			matches: true,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if tc.matches != redditAdapter.Match(&opb.Link{Href: tc.url}) {
				t.Errorf("web adapter should match %s", tc.url)
			}
		})
	}
}

func TestRedditAdapter(t *testing.T) {

	for _, tc := range []struct {
		name     string
		response string
	}{
		//{name: "basic reddit post", response: "reddut_basic"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if err := adaptertest.TestRequestResponse(
				NewAdapter(adaptertest.NewFakeHttp(filepath.Join("test_data", tc.response+".json"))),
				"http://x.com/username/status/123123123123",
				filepath.Join("test_data", tc.response+"_expect.json")); err != nil {
				t.Error(err)
			}
		})
	}
}
