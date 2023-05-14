package tokenizer

import (
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
				{Name: "tag", Params: []Param{}},
				{Text: "Text"},
				{Name: "script", Params: []Param{}},
				{Text: "the code"},
				{Name: "/script", Params: []Param{}},
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
					Params: []Param{
						{Key: "key", Value: "value"},
						{Key: "key"},
						{Key: "key", Value: "quoted value"},
						{Key: "key", Value: "also quoted"},
					},
				},
				{
					Name:   "/tag",
					Params: []Param{},
				},
			},
		},
		{
			desc:  "Escaped quotes supported",
			input: "<tag key=\"val\\\"ue\" /><tag key='val\\'ue' />",
			want: []*Token{
				{
					Name: "tag",
					Params: []Param{
						{Key: "key", Value: "val\\\"ue"},
					},
				},
				{
					Name: "tag",
					Params: []Param{
						{Key: "key", Value: "val\\'ue"},
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
