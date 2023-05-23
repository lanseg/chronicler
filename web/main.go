package main

import (
	rpb "chronicler/proto/records"
	"chronicler/util"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/url"
	"os"
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
		parent.Children = append(parent.Children, node)
		if node.Name != "" && !strings.HasPrefix(node.Name, "/") && !selfClosingTags.Contains(node.Name) {
			nodes = append([]*Node{node}, nodes...)
		}
	}
	return root
}

func main() {

	record := &rpb.Record{
		Source: &rpb.Source{
			Type: rpb.SourceType_WEB,
			Url:  os.Args[1],
		},
	}
	content, _ := ioutil.ReadFile(os.Args[1])
	record.TextContent = string(content)

	rootNode := ParseHtml(record.TextContent)
	nodes := rootNode.GetElementsByTagNames("a", "img", "script", "link")

	for _, link := range nodes {
		for _, param := range append(link.GetParam("href"), link.GetParam("src")...) {
			u, _ := url.Parse(param)
			if u.Host == "" {
				u.Host = "meduza.io"
				u.Scheme = "https"
			}
			record.Links = append(record.Links, u.String())
		}
	}
	bytes, _ := json.Marshal(record)
	fmt.Println(string(bytes))
}
