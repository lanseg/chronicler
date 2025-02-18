package twitter

import (
	"chronicler/adapter/adaptertest"
	"path/filepath"
	"testing"
)

const (
	fakeToken = "fakeToken"
)

func TestTwitterAdapter(t *testing.T) {

	for _, tc := range []struct {
		name     string
		response string
	}{
		{name: "basic tweet", response: "tweet_basic"},
		{name: "retweet", response: "retweet"},
		{name: "extended tweet", response: "tweet_extended"},
		{name: "quote retweet", response: "tweet_quote_retweet"},
		{name: "quote tweet", response: "tweet_quote"},
		{name: "tweet with media", response: "tweet_media"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if err := adaptertest.TestRequestResponse(
				NewAdapter(NewClient(
					adaptertest.NewFakeHttp(filepath.Join("test_data", tc.response+".json")),
					fakeToken)),
				"http://x.com/username/status/123123123123",
				filepath.Join("test_data", tc.response+"_expect.json")); err != nil {
				t.Error(err)
			}
		})
	}
}
