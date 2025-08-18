package twitter

import (
	"path/filepath"
	"testing"

	atest "chronicler/adapter/adaptertest"
)

const (
	fakeToken   = "fakeToken"
	twitterPost = "http://x.com/username/status/123123123123"
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
			fakeHttpClient := atest.NewFakeHttp(filepath.Join("test_data", tc.response+".json"))
			adapter := NewAdapter(fakeHttpClient, &TwitterSettings{Token: fakeToken})
			responseFile := filepath.Join("test_data", tc.response+"_expect.json")
			if err := atest.TestRequestResponse(adapter, twitterPost, responseFile); err != nil {
				t.Error(err)
			}
		})
	}
}
