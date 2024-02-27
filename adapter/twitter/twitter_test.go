package twitter

import (
	"fmt"
	"testing"

	rpb "chronicler/records/proto"
)

func newTwitterSrc(id string) []*rpb.Source {
	return []*rpb.Source{
		{
			ChannelId: id,
			Type:      rpb.SourceType_TWITTER,
		},
	}
}

func linkRecord(links ...string) *rpb.Record {
	return &rpb.Record{
		Links: links,
	}
}

func TestTwitterLinkMatcher(t *testing.T) {
	for _, tc := range []struct {
		desc   string
		record *rpb.Record
		want   []*rpb.Source
	}{
		{
			desc:   "twitter.com full link match",
			record: linkRecord("twitter.com/someAccountName/status/100500600?s=20"),
			want:   newTwitterSrc("100500600"),
		},
		{
			desc:   "x.com full link match",
			record: linkRecord("x.com/someAccountName/status/200300400500?s=20"),
			want:   newTwitterSrc("200300400500"),
		},
		{
			desc:   "x.com with prefix",
			record: linkRecord("https://x.com/someUserrr/status/200300400500"),
			want:   newTwitterSrc("200300400500"),
		},
		{
			desc:   "twitter.com with prefix",
			record: linkRecord("https://twitter.com/someUserrr/status/200300400500"),
			want:   newTwitterSrc("200300400500"),
		},
		{
			desc:   "Malformed twitter.com doesnt match",
			record: linkRecord("notAtwitter.com/someAccountName/status/400500600700?s=20"),
			want:   []*rpb.Source{},
		},
		{
			desc:   "Malformed x.com doesnt match",
			record: linkRecord("notAnx.com/someAccountName/status/200300400500?s=20"),
			want:   []*rpb.Source{},
		},
		{
			desc:   "Non-twitter path doesnt match",
			record: linkRecord("someLinkWhatever.com/x.com/twitter.com/someAccountName/status/200300400500?s=20"),
			want:   []*rpb.Source{},
		},
	} {
		t.Run(tc.desc, func(t *testing.T) {
			tg := NewTwitterAdapter(nil)

			result := tg.FindSources(tc.record)
			if fmt.Sprintf("%+v", tc.want) != fmt.Sprintf("%+v", result) {
				t.Errorf("Expected result to be:\n%+v\nBut got:\n%+v", tc.want, result)
			}
		})
	}
}
