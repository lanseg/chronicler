package util

import (
	"testing"
)

func TestLinkCheckers(t *testing.T) {
	for _, tc := range []struct {
		desc          string
		checker       func(string) bool
		skipProtocols bool
		link          string
		want          bool
	}{
		{
			desc:    "youtube regular checker success",
			checker: IsYoutubeLink,
			link:    "www.youtube.com/watch?v=dQw4w9WgXcQ&ab_channel=RickAstley",
			want:    true,
		},
		{
			desc:    "youtube short checker success",
			checker: IsYoutubeLink,
			link:    "youtu.be/dQw4w9WgXcQ",
			want:    true,
		},
		{
			desc:    "youtube mobile checker success",
			checker: IsYoutubeLink,
			link:    "m.youtu.be/watch?v=dQw4w9WgXcQ",
			want:    true,
		},
		{
			desc:    "fake youtube not matched",
			checker: IsYoutubeLink,
			link:    "www.yutube.com/watch?v=dQw4w9WgXcQ&ab_channel=RickAstley",
			want:    false,
		},
	} {
		t.Run(tc.desc, func(t *testing.T) {
			result := false
			if tc.skipProtocols {
				result = tc.checker(tc.link)
			} else {
				for _, prefix := range []string{"http://", "https://", ""} {
					result = tc.checker(prefix + tc.link)
				}
			}
			if result != tc.want {
				t.Errorf("%v(\"%v\") expected to be %v, but got %v",
					tc.checker, tc.link, tc.want, result)
			}
		})
	}
}
