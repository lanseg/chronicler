package main

import (
	"chronicler/util"
	"fmt"
	"io/ioutil"
	"net/url"
	"os"
	"runtime/pprof"
	"strings"
	"web/tokenizer"
	// rpb "chronicler/proto/records"
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

func (n *Node) GetElementsByTagName(tagName string) []*Node {
	result := []*Node{}
	toVisit := []*Node{n}
	current := n
	for len(toVisit) > 0 {
		current, toVisit = toVisit[0], toVisit[1:]
		toVisit = append(toVisit, current.Children...)
		if current.Name == tagName {
			result = append(result, current)
		}
	}
	return result
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
	f, err := os.Create("profile.prof")
	if err != nil {
		fmt.Println(err)
		return
	}
	pprof.StartCPUProfile(f)
	defer pprof.StopCPUProfile()

	log := util.NewLogger("main")
	log.Infof("Reading file")
	content, _ := ioutil.ReadFile("/home/lans/devel/chronist/sample.html")
	log.Infof("Parsing html")
	rootNode := ParseHtml(string(content))
	log.Infof("Enumerating nodes")
	nodes := rootNode.GetElementsByTagName("a")
	fmt.Println("here")

	for _, link := range nodes {
		for _, param := range link.Params {
			if param.First == "href" {
				u, _ := url.Parse(param.Second)
				if u.Host == "" {
					u.Host = "meduza.io"
					u.Scheme = "https"
				}
				fmt.Printf("%s\n", u)
			}
		}
	}
}
