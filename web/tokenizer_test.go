package tokenizer

import (
	"chronicler/util"
	"reflect"
	"testing"
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
			desc:  "Text-only content",
			input: "Some text input \nnewline\n whatever",
			want: []*Token{
				{Text: "Some text input \nnewline\n whatever"},
			},
		},
		{
			desc:  "Script special case",
			input: "<tag>Text<script>the code</script>More text",
			want: []*Token{
				{Name: "tag", Params: []util.Pair[string, string]{}},
				{Text: "Text"},
				{Name: "script", Params: []util.Pair[string, string]{}},
				{Text: "the code"},
				{Name: "/script", Params: []util.Pair[string, string]{}},
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
					Params: []util.Pair[string, string]{
						util.AsPair("key", "value"),
						util.AsPair("key", ""),
						util.AsPair("key", "quoted value"),
						util.AsPair("key", "also quoted"),
					},
				},
				{
					Name:   "/tag",
					Params: []util.Pair[string, string]{},
				},
			},
		},
		{
			desc:  "Escaped quotes supported",
			input: "<tag key=\"val\\\"ue\" /><tag key='val\\'ue' />",
			want: []*Token{
				{
					Name: "tag",
					Params: []util.Pair[string, string]{
						util.AsPair("key", "val\\\"ue"),
					},
				},
				{
					Name: "tag",
					Params: []util.Pair[string, string]{
						util.AsPair("key", "val\\'ue"),
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
