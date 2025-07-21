package common

import (
	"iter"
	"testing"
)

func TestMap(t *testing.T) {

	for _, tc := range []struct {
		name string
		data iter.Seq[int]
		want iter.Seq[string]
	}{} {
		t.Run(tc.name, func(t *testing.T) {

		})
	}
}
