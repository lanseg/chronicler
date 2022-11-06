package util

import (
	"fmt"
	"reflect"
	"testing"
)

const (
	ytLinkFull   = "www.youtube.com/watch?v=dQw4w9WgXcQ&ab_channel=RickAstley"
	ytLinkShort  = "youtu.be/dQw4w9WgXcQ"
	ytLinkMobile = "m.youtu.be/watch?v=dQw4w9WgXcQ"
	ytLinkFake   = "www.yutube.com/watch?v=dQw4w9WgXcQ&ab_channel=RickAstley"

	tgLinkShort     = "t.me/tglinkshort"
	twitterLinkPost = "twitter.com/Some_user/status/1234567890123456789"
)

func TestLinkCheckers(t *testing.T) {
	for _, tc := range []struct {
		desc    string
		checker func(string) bool
		link    string
		want    bool
	}{
		{
			desc:    "youtube regular checker success",
			checker: IsYoutubeLink,
			link:    ytLinkFull,
			want:    true,
		},
		{
			desc:    "youtube short checker success",
			checker: IsYoutubeLink,
			link:    ytLinkShort,
			want:    true,
		},
		{
			desc:    "youtube mobile checker success",
			checker: IsYoutubeLink,
			link:    ytLinkMobile,
			want:    true,
		},
		{
			desc:    "fake youtube not matched",
			checker: IsYoutubeLink,
			link:    ytLinkFake,
			want:    false,
		},
		{
			desc:    "twitter regular checker success",
			checker: IsTwitterLink,
			link:    twitterLinkPost,
			want:    true,
		},
	} {
		t.Run(tc.desc, func(t *testing.T) {
			for _, prefix := range []string{"http://", "https://", ""} {
				result := tc.checker(prefix + tc.link)
				if result != tc.want {
					t.Errorf("%v(\"%v\") expected to be %v, but got %v",
						tc.checker, tc.link, tc.want, result)
				}
			}
		})
	}
}

func TestLinkFinders(t *testing.T) {
	for _, tc := range []struct {
		desc   string
		finder func(string) []string
		text   string
		want   []string
	}{
		{
			desc:   "youtube regular finder success",
			finder: FindYoutubeLinks,
			text:   ytLinkFull,
			want:   []string{ytLinkFull},
		},
		{
			desc:   "youtube short finder success",
			finder: FindYoutubeLinks,
			text:   ytLinkShort,
			want:   []string{ytLinkShort},
		},
		{
			desc:   "youtube mobile finder success",
			finder: FindYoutubeLinks,
			text:   ytLinkMobile,
			want:   []string{ytLinkMobile},
		},
		{
			desc:   "fake youtube not matched",
			finder: FindYoutubeLinks,
			text:   ytLinkFake,
			want:   []string{},
		},
		{
			desc:   "multiple youtube links",
			finder: FindYoutubeLinks,
			text: fmt.Sprintf("Lorem %s ipsum dolor %s %s sit amet, consectetur adipiscing elit,"+
				"sed do eiusmod tempor %s incididunt ut labore et dolore magna aliqua.",
				ytLinkFull, ytLinkFake, ytLinkMobile, ytLinkShort),
			want: []string{ytLinkFull, ytLinkMobile, ytLinkShort},
		},
		{
			desc:   "multiple youtube links web finder",
			finder: FindWebLinks,
			text: fmt.Sprintf("Lorem %s ipsum dolor %s %s sit amet, consectetur adipiscing elit,"+
				"sed do eiusmod tempor %s incididunt ut labore et dolore magna aliqua.",
				ytLinkFull, ytLinkFake, ytLinkMobile, ytLinkShort),
			want: []string{ytLinkFull, ytLinkFake, ytLinkMobile, ytLinkShort},
		},
		{
			desc:   "multiple links web finder",
			finder: FindWebLinks,
			text: fmt.Sprintf("Lorem %s ipsum dolor %s %s sit amet, consectetur adipiscing elit,"+
				"sed do eiusmod tempor %s incididunt ut labore et dolore magna aliqua.",
				ytLinkFull, tgLinkShort, ytLinkMobile, twitterLinkPost),
			want: []string{ytLinkFull, tgLinkShort, ytLinkMobile, twitterLinkPost},
		},
		{
			desc:   "multiple links with prefixes",
			finder: FindWebLinks,
			text: fmt.Sprintf("Lorem http://%s ipsum dolor https://%s %s sit amet, consectetur adipiscing elit,"+
				"sed do eiusmod tempor https://%s incididunt ut labore et dolore magna aliqua.",
				ytLinkFull, tgLinkShort, ytLinkMobile, twitterLinkPost),
			want: []string{"http://" + ytLinkFull, "https://" + tgLinkShort, ytLinkMobile, "https://" + twitterLinkPost},
		},
		{
			desc:   "Find twitter id full link",
			finder: FindTwitterIds,
			text:   "twitter.com/Some_user/status/1234567890123456789?qwodmqwodm1=123",
			want: []string{
				"1234567890123456789",
			},
		},
	} {
		t.Run(tc.desc, func(t *testing.T) {
			result := tc.finder(tc.text)
			if (len(result) != 0 || len(tc.want) != 0) && !reflect.DeepEqual(result, tc.want) {
				t.Errorf("%v(\"%v\") expected to be %v, but got %v",
					tc.finder, tc.text, tc.want, result)
			}
		})
	}
}
