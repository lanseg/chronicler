package twitter

import (
	"fmt"
	"testing"

	rpb "chronicler/records/proto"
)

func newTwitterSrc(id string) *rpb.Source {
	return &rpb.Source{
		ChannelId: id,
		Type:      rpb.SourceType_TWITTER,
	}
}

func TestTwitterLinkMatcher(t *testing.T) {
	for _, tc := range []struct {
		desc string
		link string
		want *rpb.Source
	}{
		{
			desc: "twitter.com full link match",
			link: "twitter.com/someAccountName/status/100500600?s=20",
			want: newTwitterSrc("100500600"),
		},
		{
			desc: "x.com full link match",
			link: "x.com/someAccountName/status/200300400500?s=20",
			want: newTwitterSrc("200300400500"),
		},
		{
			desc: "x.com with prefix",
			link: "https://x.com/someUserrr/status/200300400500",
			want: newTwitterSrc("200300400500"),
		},
		{
			desc: "twitter.com with prefix",
			link: "https://twitter.com/someUserrr/status/200300400500",
			want: newTwitterSrc("200300400500"),
		},
		{
			desc: "Malformed twitter.com doesnt match",
			link: "notAtwitter.com/someAccountName/status/400500600700?s=20",
			want: nil,
		},
		{
			desc: "Malformed x.com doesnt match",
			link: "notAnx.com/someAccountName/status/200300400500?s=20",
			want: nil,
		},
		{
			desc: "Non-twitter path doesnt match",
			link: "someLinkWhatever.com/x.com/twitter.com/someAccountName/status/200300400500?s=20",
			want: nil,
		},
	} {
		t.Run(tc.desc, func(t *testing.T) {
			tg := NewTwitterAdapter(nil)

			result := tg.MatchLink(tc.link)
			if fmt.Sprintf("%+v", tc.want) != fmt.Sprintf("%+v", result) {
				t.Errorf("Expected result to be:\n%+v\nBut got:\n%+v", tc.want, result)
			}
		})
	}
}
