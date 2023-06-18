package htmlparser

import (
	"chronicler/util"
	"strings"
	"web/tokenizer"

	"github.com/lanseg/golang-commons/collections"
)

var (
	selfClosingTags = util.NewSet([]string{
		"area", "base", "br", "col", "embed",
		"hr", "img", "input", "link", "meta",
		"param", "source", "track", "wbr",
	})
)

func getChildren(n *Node) []*Node {
	if n.Children == nil {
		return []*Node{}
	}
	return n.Children
}

type Node struct {
	Name     string
	Text     string
	Params   map[string][]string
	Parent   *Node
	Children []*Node
}

func (n *Node) GetParam(paramName string) []string {
	if value, ok := n.Params[paramName]; ok {
		return value
	}
	return []string{}
}

func (n *Node) GetElementsByTagNames(tagNames ...string) []*Node {
	names := util.NewSet(tagNames)
	return collections.IterateTree(n, getChildren).Filter(
		func(n *Node) bool {
			return names.Contains(n.Name)
		}).Collect()
}

func newNode(t *tokenizer.Token) *Node {
	params := map[string][]string{}
	for _, param := range t.Params {
		if _, ok := params[param.First]; !ok {
			params[param.First] = []string{}
		}
		params[param.First] = append(params[param.First], param.Second)
	}
	return &Node{
		Name:     t.Name,
		Text:     t.Text,
		Params:   params,
		Children: []*Node{},
	}
}

func ParseHtml(content string) *Node {
	tokens := tokenizer.Tokenize(content)
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
		node.Parent = parent
		parent.Children = append(parent.Children, node)
		if node.Name != "" && !strings.HasPrefix(node.Name, "/") && !selfClosingTags.Contains(node.Name) {
			nodes = append([]*Node{node}, nodes...)
		}
	}
	return root
}

func GetTitle(content string) string {
	doc := ParseHtml(content)
	result := strings.Builder{}
	collections.IterateTree(doc, getChildren).ForEachRemaining(
		func(n *Node) bool {
			if n.Parent != nil && n.Parent.Name == "title" {
				result.WriteString(n.Text)
				return true
			}
			return false
		})
	return result.String()
}

func StripTags(content string) string {
	doc := ParseHtml(content)
	result := strings.Builder{}
	excludedLinks := util.NewSet([]string{
		"script", "meta", "link", "head", "style",
	})
	collections.IterateTree(doc, getChildren).Filter(
		func(n *Node) bool {
			if n.Parent == nil || (n.Name == "" && n.Text == "") {
				return false
			}
			return !excludedLinks.Contains(n.Parent.Name) && !excludedLinks.Contains(n.Name)
		}).ForEachRemaining(
		func(n *Node) bool {
			result.WriteString(n.Text)
			result.WriteRune(' ')
			return false
		})
	return result.String()
}
