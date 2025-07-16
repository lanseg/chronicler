package common

import (
	"os"
	"reflect"
	"regexp"
	"strings"
	"testing"
)

func TestWrapText(t *testing.T) {
	data, err := os.ReadFile("test_data/lipsum.txt")
	if err != nil {
		t.Errorf("Cannot load test data: %s", err)
	}

	lipsum := string(data)

	t.Run("Simple test", func(t *testing.T) {
		maxWidth := 20
		lipsumLines := strings.Split(WrapText(lipsum, maxWidth), "\n")
		for _, l := range lipsumLines {
			if len(l) > maxWidth {
				t.Errorf("Line is longer than max %d characters: %q", maxWidth, l)
				break
			}
		}
	})

	t.Run("One symbol width test", func(t *testing.T) {
		maxWidth := 1
		splitter := regexp.MustCompile(`\s+`)
		lipsumLines := strings.Split(WrapText(lipsum, maxWidth), "\n")
		for _, l := range lipsumLines {
			words := splitter.Split(l, -1)
			if len(words) > 2 {
				t.Errorf("Unexpected word in line %q: %q", l, words)
				break
			}
		}
	})

	t.Run("width out of bounds", func(t *testing.T) {
		zeroWidth := WrapText(lipsum, 0)
		oneSplit := WrapText(lipsum, 1)
		negativeSplit := WrapText(lipsum, -1)
		if !reflect.DeepEqual(zeroWidth, negativeSplit) {
			t.Errorf("Negative or zero split means just one word per line, but got %q and %q",
				zeroWidth, negativeSplit)
		}
		if !reflect.DeepEqual(negativeSplit, oneSplit) {
			t.Errorf("1 split or zero split means just one word per line, but got %q and %q",
				zeroWidth, negativeSplit)
		}
	})

	t.Run("wrap empty string", func(t *testing.T) {
		zeroWrap := WrapText("", 0)
		if zeroWrap != "" {
			t.Errorf("Wrapping empty string should return empty string, but got %q", zeroWrap)
		}

		tenWrap := WrapText("", 10)
		if tenWrap != "" {
			t.Errorf("Wrapping empty string should return empty string, but got %q", tenWrap)
		}
	})

	t.Run("wrap single word", func(t *testing.T) {
		srcStr := "abcdef"
		result := WrapText(srcStr, 0)
		if result != srcStr {
			t.Errorf("Wrapping %q should return %q, but got %q", srcStr, srcStr, result)
		}
	})
}

func TestSanitize(t *testing.T) {

	for _, tc := range []struct {
		name   string
		url    string
		want   string
		maxLen int
	}{
		{
			name: "basic url",
			url:  "http://somehost.domain.com/query?a=b&c=d",
			want: "http___somehost.domain.com_query_a_b_c_d_d3a8884e",
		},
		{
			name:   "basic url limited not truncated",
			url:    "http://somehost.domain.com/query?a=b&c=d",
			want:   "http___somehost.domain.com_query_a_b_c_d_d3a8884e",
			maxLen: 320,
		},
		{
			name:   "basic url limited truncated",
			url:    "http://somehost.domain.com/query?a=b&c=d",
			want:   "http___somehost.domain._d3a8884e",
			maxLen: 32,
		},
		{
			name: "basic url unicode",
			url:  "http://somehost.domain.com/query?a=b&wwwwtpowkfwüokзщулкпзщ3",
			want: "http___somehost.domain.com_query_a_b_wwwwtpowkfw_ok________3_0a80a2de",
		},
		{
			name: "empty string",
			url:  "",
			want: "",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			result := Sanitize(tc.url, tc.maxLen)
			if tc.want != result || (tc.maxLen > 0 && tc.maxLen < len(result)) {
				t.Errorf("Expected Sanitize(%q, %d)=%q (%d), but got %q (%d)",
					tc.url, tc.maxLen, tc.want, tc.maxLen, result, len(result))
			}
		})
	}

	for _, tc := range []struct {
		name          string
		urlA          string
		urlB          string
		maxLen        int
		wantCollision bool
	}{
		{
			name:          "different urls different sanitized",
			urlA:          "http://somehost.domain.com/query?a=b&c=d",
			urlB:          "http://somehost.domain.com/query?a=d&b=c",
			wantCollision: false,
		},
		{
			name:          "different urls same sanitized",
			urlA:          "http://somehost.domain.com/query?А=Б&Ц=Д",
			urlB:          "http://somehost.domain.com/query?Б=А&Д=Ц",
			wantCollision: false,
		},
		{
			name:          "different urls same after truncation",
			urlA:          "http://somehost.domain.com/query/some/long/path?А=Б&Ц=Д",
			urlB:          "http://somehost.domain.com/query/some/long/path?Б=А&Д=Ц",
			maxLen:        20,
			wantCollision: false,
		},
		{
			name:          "same urls same sanitized after truncation",
			urlA:          "http://somehost.domain.com/query?А=Б&Ц=Д",
			urlB:          "http://somehost.domain.com/query?Б=А&Д=Ц",
			maxLen:        20,
			wantCollision: false,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			resultA := Sanitize(tc.urlA, tc.maxLen)
			resultB := Sanitize(tc.urlB, tc.maxLen)
			if resultA == resultB && !tc.wantCollision {
				t.Errorf("Expected Sanitize(%q, %d)=%q, Sanitize(%q, %d)=%q to be different, but "+
					"they are the same",
					tc.urlA, tc.maxLen, resultA, tc.urlB, tc.maxLen, resultB)
			}
		})
	}
}

func TestSanitizeWithExt(t *testing.T) {

	for _, tc := range []struct {
		name   string
		url    string
		mime   string
		want   string
		maxLen int
	}{
		{
			name: "basic url without extension",
			url:  "http://somehost.domain.com/query?a=b&c=d",
			want: "http___somehost.domain.com_query_a_b_c_d_d3a8884e",
		},
		{
			name: "basic url with extension",
			url:  "http://somehost.domain.com/path/to/file.png?a=b&c=d",
			want: "http___somehost.domain.com_path_to_file.png_a_b_c_d_676dc64b.png",
		},
		{
			name:   "basic url limited not truncated",
			url:    "http://somehost.domain.com/query?a=b&c=d",
			want:   "http___somehost.domain.com_query_a_b_c_d_d3a8884e",
			maxLen: 320,
		},
		{
			name:   "basic url limited truncated",
			url:    "http://somehost.domain.com/query?a=b&c=d",
			want:   "http___somehost.domain._d3a8884e",
			maxLen: 32,
		},
		{
			name:   "basic url with extension limited truncated",
			url:    "http://somehost.domain.com/path/to/file.png?query&a=b&c=d",
			want:   "http___somehost.dom_41afbf31.png",
			maxLen: 32,
		},
		{
			name: "basic url unicode",
			url:  "http://somehost.domain.com/query?a=b&wwwwtpowkfwüokзщулкпзщ3",
			want: "http___somehost.domain.com_query_a_b_wwwwtpowkfw_ok________3_0a80a2de",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			result := SanitizeWithExt(tc.url, tc.mime, tc.maxLen)
			if tc.want != result || (tc.maxLen > 0 && tc.maxLen < len(result)) {
				t.Errorf("Expected Sanitize(%q, %d)=%q (%d), but got %q (%d)",
					tc.url, tc.maxLen, tc.want, tc.maxLen, result, len(result))
			}
		})
	}
}
