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
			url:     "https://www.reddit.com/r/subreddit/comments/1fsokfgg/mulfoe/",
			matches: true,
		},
		{
			name:    "comment succesful match",
			url:     "https://www.reddit.com/r/subreddit/comments/1fsokfgg/comment/mc9kwefuo5g/?utm_source=share&utm_medium=web3x&utm_name=web3xcss&utm_term=1&utm_content=share_button",
			matches: true,
		},
		{
			name:    "failed match subreddit only",
			url:     "https://www.reddit.com/r/subreddit",
			matches: false,
		},
		{
			name:    "failed match no post id",
			url:     "https://www.reddit.com/r/subreddit/comments",
			matches: false,
		},
		{
			name:    "failed match garbage url",
			url:     "eg=elkrg[lreg=3rgergr/lasubredditw",
			matches: false,
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
		{name: "reddit basic short post", response: "basic_post"},
		{name: "reddit basic post with video", response: "video_post"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if err := adaptertest.TestRequestResponse(
				NewAdapter(
					adaptertest.NewFakeHttp(filepath.Join("test_data", tc.response+".json"))),
				"https://www.reddit.com/r/subreddit/comments/rand0m/comment/mc9uo5u",
				filepath.Join("test_data", tc.response+"_expect.json")); err != nil {
				t.Error(err)
			}
		})
	}
}
