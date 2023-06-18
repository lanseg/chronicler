package tokenizer

import (
	"reflect"
	"testing"

	"github.com/lanseg/golang-commons/collections"
)

func TestTwitterClient(t *testing.T) {

	for _, tc := range []struct {
		desc  string
		input string
		want  []*Token
	}{
		{
			desc:  "Text-only content",
			input: "Some text input \nnewline\n whatever",
			want: []*Token{
				{Text: "Some text input \nnewline\n whatever"},
			},
		},
		{
			desc:  "Text-only content without newlines unicode",
			input: "Текст в юникоде 中國 भारत 한국",
			want: []*Token{
				{Text: "Текст в юникоде 中國 भारत 한국"},
			},
		},
		{
			desc:  "Script special case",
			input: "<tag>Text<script>the code</script>More text",
			want: []*Token{
				{Name: "tag", Params: []collections.Pair[string, string]{}},
				{Text: "Text"},
				{Name: "script", Params: []collections.Pair[string, string]{}},
				{Text: "the code"},
				{Name: "/script", Params: []collections.Pair[string, string]{}},
				{Text: "More text"},
			},
		},
		{
			desc:  "Single tag content",
			input: "Some input <tag key=value key key=\"quoted value\" key='also quoted'></tag>",
			want: []*Token{
				{Text: "Some input "},
				{
					Name: "tag",
					Params: []collections.Pair[string, string]{
						collections.AsPair("key", "value"),
						collections.AsPair("key", ""),
						collections.AsPair("key", "quoted value"),
						collections.AsPair("key", "also quoted"),
					},
				},
				{
					Name:   "/tag",
					Params: []collections.Pair[string, string]{},
				},
			},
		},
		{
			desc:  "Params with dash",
			input: "<tag data-value=\"some value\" />",
			want: []*Token{
				{
					Name: "tag",
					Params: []collections.Pair[string, string]{
						collections.AsPair("data-value", "some value"),
					},
				},
			},
		},
		{
			desc:  "Escaped quotes supported",
			input: "<tag key=\"val\\\"ue\" /><tag key='val\\'ue' />",
			want: []*Token{
				{
					Name: "tag",
					Params: []collections.Pair[string, string]{
						collections.AsPair("key", "val\\\"ue"),
					},
				},
				{
					Name: "tag",
					Params: []collections.Pair[string, string]{
						collections.AsPair("key", "val\\'ue"),
					},
				},
			},
		},
	} {
		t.Run(tc.desc, func(t *testing.T) {
			result := Tokenize(tc.input)
			if !reflect.DeepEqual(result, tc.want) {
				t.Errorf("Tokenize(%s) expected to be %s, but got %s",
					tc.input, tc.want, result)
			}
		})
	}
}
