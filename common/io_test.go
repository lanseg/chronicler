package common

import (
	"testing"
)

func TestGuessType(t *testing.T) {

	for _, tc := range []struct {
		name string
		url  string
		want string
	}{
		{"link with text file", "http://whatever.com/path/to/file.txt", "text/plain; charset=utf-8"},
		{"link with image file", "http://whatever.com/path/to/file.jpg", "image/jpeg"},
		{"link with gzip file", "http://whatever.com/path/to/file.tar.gz", "application/gzip"},
		{"link without file", "http://whatever.com", ""},
		{"broken link returns empty type", "wh  at", ""},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got := GuessMimeType(tc.url)
			if got != tc.want {
				t.Errorf("Expected %q to have type %q, but got %q", tc.url, tc.want, got)
			}
		})
	}
}
