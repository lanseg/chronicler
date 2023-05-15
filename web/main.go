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

type Node struct {
	Name     string
	Text     string
	Params   []util.Pair[string, string]
	Children []*Node
}

func (n *Node) dump(prefix string) {
	if n.Name == "" {
		fmt.Printf("%s%s\n", prefix, n.Text)
	} else {
		fmt.Printf("%s[%s]\n", prefix, n.Name)
	}
	for _, child := range n.Children {
		child.dump(prefix + "  ")
	}
}

func newNode(t *tokenizer.Token) *Node {
	return &Node{
		Name:     t.Name,
		Text:     t.Text,
		Params:   t.Params,
		Children: []*Node{},
	}
}

func ParseHtml(content string) *Node {
	tokens := tokenizer.Tokenize(content)
	fmt.Printf("Tokens: %d, %s\n", len(tokens), tokens[len(tokens)-1])
	root := &Node{
		Name:     "#root",
		Children: []*Node{},
	}
	nodes := []*Node{root}
	for _, token := range tokens {
		parent := nodes[0]
		if ("/" + parent.Name) == token.Name {
			nodes = nodes[1:]
			continue
		}
		node := newNode(token)
		parent.Children = append(parent.Children, node)
		if node.Name != "" && !strings.HasPrefix(node.Name, "/") && !selfClosingTags.Contains(node.Name) {
			nodes = append([]*Node{node}, nodes...)
		}
	}
	return root
}

func main() {
	fmt.Println("Reading file")
	content, _ := ioutil.ReadFile("/home/lans/devel/chronist/sample.html")
	fmt.Println("Parsing file")
	rootNode := ParseHtml(string(content))
	rootNode.dump("")
}
