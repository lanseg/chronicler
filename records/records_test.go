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

func TestMergeStrings(t *testing.T) {

	for _, tc := range []struct {
		desc string
		a    []string
		b    []string
		want []string
	}{
		{
			desc: "Two nils, return empty result ",
			a:    nil,
			b:    nil,
			want: []string{},
		},
		{
			desc: "A nil, b non nil returns b",
			a:    nil,
			b:    []string{"D", "E", "F", "A", "C", "B", "F", "A"},
			want: []string{"A", "B", "C", "D", "E", "F"},
		},
		{
			desc: "A non nil, b nil returns a",
			a:    []string{"D", "E", "F", "A", "C", "B"},
			b:    nil,
			want: []string{"A", "B", "C", "D", "E", "F"},
		},
		{
			desc: "Both non nil, returns merged and sorted",
			a:    []string{"A", "D", "B", "F", "H"},
			b:    []string{"E", "A", "C", "D", "H"},
			want: []string{"A", "B", "C", "D", "E", "F", "H"},
		},
	} {
		t.Run(tc.desc, func(t *testing.T) {
			result := MergeStrings(tc.a, tc.b)
			if !reflect.DeepEqual(result, tc.want) {
				t.Errorf("MergeStrings(%s, %s) expected to be %s, but got %s",
					tc.a, tc.b, tc.want, result)
			}
		})
	}
}

func TestMergeUserMetadata(t *testing.T) {

	for _, tc := range []struct {
		desc string
		a    []*rpb.UserMetadata
		b    []*rpb.UserMetadata
		want []*rpb.UserMetadata
	}{
		{
			desc: "Two nils return empty",
			a:    nil,
			b:    nil,
			want: []*rpb.UserMetadata{},
		},
		{
			desc: "A not nil, b nil returns a",
			a:    []*rpb.UserMetadata{{Id: "123", Username: "456", Quotes: []string{"a", "b", "c"}}},
			b:    nil,
			want: []*rpb.UserMetadata{{Id: "123", Username: "456", Quotes: []string{"a", "b", "c"}}},
		},
		{
			desc: "B not nil, a nil returns b",
			a:    nil,
			b:    []*rpb.UserMetadata{{Id: "123", Username: "456", Quotes: []string{"a", "b", "c"}}},
			want: []*rpb.UserMetadata{{Id: "123", Username: "456", Quotes: []string{"a", "b", "c"}}},
		},
		{
			desc: "Both not nil",
			a: []*rpb.UserMetadata{
				{Id: "123456", Username: "789", Quotes: []string{"a", "b", "c"}},
				{Id: "12345", Username: "", Quotes: []string{"o", "m", "g"}},
			},
			b: []*rpb.UserMetadata{
				{Id: "12345", Username: "456", Quotes: []string{"d", "e", "f"}},
				{Id: "123456", Username: "", Quotes: []string{"g", "h", "i"}},
			},
			want: []*rpb.UserMetadata{
				{Id: "12345", Username: "456", Quotes: []string{"d", "e", "f", "g", "m", "o"}},
				{Id: "123456", Username: "789", Quotes: []string{"a", "b", "c", "g", "h", "i"}}},
		},
	} {
		t.Run(tc.desc, func(t *testing.T) {
			result := MergeUserMetadata(tc.a, tc.b)
			if !reflect.DeepEqual(result, tc.want) {
				t.Errorf("MergeUserMetadata(%s, %s) expected to be %s, but got %s",
					tc.a, tc.b, tc.want, result)
			}
		})
	}
}
