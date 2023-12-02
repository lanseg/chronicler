package tinyhtml

import (
	"fmt"
	"sort"
	"strings"

	"github.com/lanseg/golang-commons/collections"
	"github.com/lanseg/golang-commons/optional"
)

var (
	selfClosingTags = collections.NewSet([]string{
		"area", "base", "br", "col", "embed",
		"hr", "img", "input", "link", "meta",
		"param", "source", "track", "wbr",
		"!DOCTYPE", "#text",
	})
	emptyParam = &optional.Nothing[string]{}
)

func getChildren(n *Node) []*Node {
	if n.Children == nil {
		return []*Node{}
	}
	return n.Children
}

type RawFragment struct {
	Start int
	End   int
	Raw   string
}

type Node struct {
	Name     string
	Value    string
	Raw      *RawFragment
	Params   map[string]string
	Children []*Node
}

func (n *Node) iterateChildren() collections.Iterator[*Node] {
	return collections.IterateTree(n, collections.DepthFirst, func(node *Node) []*Node {
		if node.Children == nil {
			return []*Node{}
		}
		return node.Children
	})
}

func (n *Node) InnerHTML() string {
	result := strings.Builder{}
	n.iterateChildren().ForEachRemaining(func(node *Node) bool {
		result.WriteString(node.Raw.Raw)
		return false
	})
	return result.String()
}

func (n *Node) GetElementsByTagAndClass(tag string, classes ...string) []*Node {
	return n.iterateChildren().Filter(func(node *Node) bool {
		data, ok := node.Params["class"]
		if tag != node.Name || (!ok && len(classes) > 0) {
			return false
		}
		return collections.NewSet(strings.Split(data, " ")).ContainsAll(classes)
	}).Collect()
}

func (n *Node) GetElementsByTags(tags ...string) []*Node {
	tagSet := collections.NewSet(tags)
	return n.iterateChildren().Filter(func(node *Node) bool {
		return tagSet.Contains(node.Name)
	}).Collect()
}

func (n *Node) GetAttribute(attr string) optional.Optional[string] {
	if value, ok := n.Params[attr]; ok {
		return optional.Of(value)
	}
	return emptyParam
}

func (n *Node) String() string {
	params := []string{}
	for k, v := range n.Params {
		params = append(params, fmt.Sprintf("%q=%q", k, v))
	}
	sort.Strings(params)
	return fmt.Sprintf("Node { %q %s}", n.Name, strings.Join(params, ", "))
}

func NewNode(name string) *Node {
	return &Node{
		Name:     name,
		Params:   map[string]string{},
		Raw:      &RawFragment{},
		Children: []*Node{},
	}
}

func NewDataNode(data string) *Node {
	return &Node{
		Name:     "#text",
		Params:   map[string]string{},
		Value:    data,
		Raw:      &RawFragment{},
		Children: []*Node{},
	}
}

func ParseHtml(doc string) (*Node, error) {
	lastParam := ""
	nodes := []*Node{}
	runes := []rune(doc)
	for _, t := range tokenize(doc) {
		switch t.tokenType {
		case DATA:
			node := NewDataNode(t.value)
			node.Raw = &RawFragment{
				Start: t.start,
				End:   t.end,
				Raw:   string(runes[t.start:t.end]),
			}
			if len(nodes) > 0 {
				prevRaw := nodes[len(nodes)-1].Raw
				prevRaw.End = t.start
				prevRaw.Raw = string(runes[prevRaw.Start:prevRaw.End])
			}
			nodes = append(nodes, node)
		case TAG_NAME:
			node := NewNode(t.value)
			node.Raw = &RawFragment{
				Start: t.start,
				End:   t.end,
				Raw:   string(runes[t.start:t.end]),
			}
			prevRaw := nodes[len(nodes)-1].Raw
			prevRaw.End = t.start
			prevRaw.Raw = string(runes[prevRaw.Start:prevRaw.End])
			nodes = append(nodes, node)
		case TAG_PARAM:
			nodes[len(nodes)-1].Params[t.value] = ""
			lastParam = t.value
		case TAG_VALUE:
			nodes[len(nodes)-1].Params[lastParam] = t.value
		}
	}

	root := NewNode("#root")
	stack := []*Node{root}
	for _, n := range nodes {
		if n.Name[0] == '/' {
			stack = stack[:len(stack)-1]
			continue
		}
		parent := stack[len(stack)-1]
		parent.Children = append(parent.Children, n)
		if !selfClosingTags.Contains(n.Name) {
			stack = append(stack, n)
		}
	}
	return root, nil
}

func dump(n *Node, prefix string) {
	fmt.Printf("%s %s\n", prefix, n)
	for _, nn := range n.Children {
		dump(nn, prefix+"  ")
	}
}

func GetTitle(s string) string {
	nodes, _ := ParseHtml(s)
	titles := nodes.GetElementsByTagAndClass("title")
	if len(titles) == 0 || len(titles[0].Children) == 0 {
		return ""
	}
	return titles[0].Children[0].Value
}

func StripTags(s string) string {
	nodes, _ := ParseHtml(s)
	result := strings.Builder{}
	nodes.iterateChildren().ForEachRemaining(func(n *Node) bool {
		if n.Name != "#text" || n.Value == "" {
			return false
		}
		result.WriteString(n.Value)
		return false
	})
	return result.String()
}
