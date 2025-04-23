package common

import (
	"net/url"
	"testing"
)

func TestNormalizeURL(t *testing.T) {
	for _, tc := range []struct {
		name string
		url  string
		want string
	}{
		{"percent uppercase", "http://example.com/foo%2a", "http://example.com/foo*"},
		{"percent space", "http://example.com/foo%20bar", "http://example.com/foo%20bar"},
		{"scheme and host to lower", "HTTP://User@Example.COM/Foo", "http://User@example.com/Foo"},
		{"default scheme added", "//example.com", "http://example.com/"},
		{"decode unneeded percent", "http://example.com/%7Efoo", "http://example.com/~foo"},
		{"remove dot segments", "http://example.com/foo/./bar/baz/../qux", "http://example.com/foo/bar/qux"},
		{"empty to slash", "http://example.com", "http://example.com/"},
		{"remove default http port", "http://example.com:80/", "http://example.com/"},
		{"remove default https port", "https://example.com:443/", "https://example.com/"},
		{"remove fragment", "http://example.com/bar.html#section1", "http://example.com/bar.html"},
		{"remove duplicate slashes", "http://example.com/foo//bar.html", "http://example.com/foo/bar.html"},
		{"remove unneeded question mark", "http://example.com/display?", "http://example.com/display"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			url, err := url.Parse(tc.url)
			result := NormalizeURL(url)
			if err != nil {
				t.Errorf("expected normalizeUrl(%q) == %q, but got %q",
					tc.url, tc.want, err)
				return
			}
			if url.String() != tc.want {
				t.Errorf("expected normalizeUrl(%q) == %q, but got %q",
					tc.url, tc.want, result.String())
			}
		})
	}
}
