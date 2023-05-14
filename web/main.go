package main

import (
	"chronicler/util"
	"fmt"
	"io/ioutil"
	"strings"
	"web/tokenizer"
)

var (
	selfClosingTags = util.NewSet([]string{
		"area", "base", "br", "col", "embed",
		"hr", "img", "input", "link", "meta",
		"param", "source", "track", "wbr",
	})
)

type Tag struct {
	Name     string
	Token    *tokenizer.Token
	Children []*Tag
}

func (t *Tag) dump(prefix string) {
	if t.Token.Name == "" {
		fmt.Printf("%s%s\n", prefix, t.Token.Text)
	} else {
		fmt.Printf("%s[%s]\n", prefix, t.Token.Name)
	}
	for _, child := range t.Children {
		child.dump(prefix + "  ")
	}
}

func newTag(t *tokenizer.Token) *Tag {
	return &Tag{
		Token:    t,
		Children: []*Tag{},
	}
}

func ParseHtml(content string) *Tag {
	tokens := tokenizer.Tokenize(content)
	fmt.Printf("Tokens: %d, %s\n", len(tokens), tokens[len(tokens)-1])
	root := &Tag{
		Name:     "root",
		Token:    &tokenizer.Token{Name: "root"},
		Children: []*Tag{},
	}
	tagStack := []*Tag{root}
	for _, token := range tokens {
		parent := tagStack[0]
		if ("/" + parent.Name) == token.Name {
			tagStack = tagStack[1:]
			continue
		}
		newTag := newTag(token)
		parent.Children = append(parent.Children, newTag)
		if newTag.Name != "" && !strings.HasPrefix(newTag.Name, "/") && !selfClosingTags.Contains(newTag.Name) {
			tagStack = append([]*Tag{newTag}, tagStack...)
		}
	}
	return root
}

func main() {
	fmt.Println("Reading file")
	content, _ := ioutil.ReadFile("/home/lans/devel/chronist/sample.html")
	fmt.Println("Parsing file")
	tag := ParseHtml(string(content))
	tag.dump("")
}
