package tinyhtml

import (
	"os"
	"testing"
)

func TestTinyHtml(t *testing.T) {

	for _, tc := range []struct {
		name string
		path string
	}{
		{
			name: "Test some html",
			path: "testdata/basic_valid.html",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			file, err := os.ReadFile(tc.path)
			if err != nil {
				t.Errorf("Error: %s", err)
			}
			_, err = ParseHtml(string(file))
			if err != nil {
				t.Errorf("Error while parsing html: %s", err)
			}
		})
	}
}
