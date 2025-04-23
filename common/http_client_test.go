package common

import (
	"testing"
)

func TestHttpClient(t *testing.T) {
	for _, tc := range []struct {
		name     string
		settings HttpSettings
		want     HttpClient
	}{} {
		t.Run(tc.name, func(t *testing.T) {
		})
	}
}
