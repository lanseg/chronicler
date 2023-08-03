package records

import (
	"reflect"
	"testing"

	rpb "chronicler/records/proto"
)

func TestMergeFiles(t *testing.T) {

	for _, tc := range []struct {
		desc string
		a    []*rpb.File
		b    []*rpb.File
		want []*rpb.File
	}{
		{
			desc: "Two nils return empty",
			a:    nil,
			b:    nil,
			want: []*rpb.File{},
		},
		{
			desc: "A not nil, b nil returns a",
			a:    []*rpb.File{{FileId: "123", FileUrl: "456"}},
			b:    nil,
			want: []*rpb.File{{FileId: "123", FileUrl: "456"}},
		},
		{
			desc: "B not nil, a nil returns b",
			a:    nil,
			b:    []*rpb.File{{FileId: "12345", FileUrl: "456"}},
			want: []*rpb.File{{FileId: "12345", FileUrl: "456"}},
		},
		{
			desc: "Both not nil",
			a: []*rpb.File{
				{FileId: "123456", FileUrl: "789"},
			},
			b: []*rpb.File{
				{FileId: "12345", FileUrl: "456"},
			},
			want: []*rpb.File{
				{FileId: "12345", FileUrl: "456"},
				{FileId: "123456", FileUrl: "789"},
			},
		},
	} {
		t.Run(tc.desc, func(t *testing.T) {
			result := MergeFiles(tc.a, tc.b)
			if !reflect.DeepEqual(result, tc.want) {
				t.Errorf("MergeFiles(%s, %s) expected to be %s, but got %s",
					tc.a, tc.b, tc.want, result)
			}
		})
	}
}
