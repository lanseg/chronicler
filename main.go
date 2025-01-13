package main

import (
	"fmt"
	"io"
	"os"
	"slices"
	"strconv"
	"strings"
	"time"

	"golang.org/x/net/html"
	"golang.org/x/text/encoding/charmap"

	opb "chronicler/proto"
)

const (
	InDocument       = 0
	InArticle        = 1
	InArticleTitle   = 2
	InComment        = 3
	InCommentContent = 4
)

func toTimestamp(s string) (*opb.Timestamp, error) {
	i, err := strconv.Atoi(s)
	if err == nil {
		return &opb.Timestamp{Seconds: int64(i)}, nil
	}
	t, err := time.Parse(time.RFC3339, s)
	if err == nil {
		return &opb.Timestamp{Seconds: t.Unix(), Nanos: int32(t.Nanosecond())}, nil
	}
	return nil, err
}

func attrMap(token *html.Token) map[string]string {
	result := map[string]string{}
	for _, attr := range token.Attr {
		result[attr.Key] = attr.Val
	}
	return result
}

func hasClass(token *html.Token, class string) bool {
	for _, attr := range token.Attr {
		if attr.Key == "class" && slices.Contains(strings.Split(attr.Val, " "), class) {
			return true
		}
	}
	return false
}

func getAttachment(token *html.Token) *opb.Attachment {
	if token.Type != html.StartTagToken && token.Type != html.SelfClosingTagToken {
		return nil
	}
	attrs := attrMap(token)
	links := []string{}
	if attrs["href"] != "" {
		links = append(links, attrs["href"])
	}
	if attrs["src"] != "" {
		links = append(links, attrs["src"])
	}
	if attrs["data-src"] != "" {
		links = append(links, attrs["data-src"])
	}
	if len(links) == 0 {
		return nil
	}
	return &opb.Attachment{
		Link: links,
	}
}

type PikabuParserSM struct {
	tokenizer *html.Tokenizer
	token     *html.Token
	attrs     map[string]string
	state     int

	article *opb.Object
	comment []*opb.Object
}

func (psm *PikabuParserSM) nextToken() bool {
	if tokenType := psm.tokenizer.Next(); tokenType == html.ErrorToken {
		return true
	}
	token := psm.tokenizer.Token()
	psm.token = &token
	if psm.attrs == nil || len(psm.attrs) != 0 {
		psm.attrs = map[string]string{}
	}
	return false
}

func (psm *PikabuParserSM) attr(key string) string {
	if len(psm.token.Attr) != 0 && len(psm.attrs) == 0 {
		psm.attrs = attrMap(psm.token)
	}
	if value, ok := psm.attrs[key]; ok {
		return value
	}
	return ""
}

func (psm *PikabuParserSM) newArticle() {
	token := psm.token
	attrs := attrMap(token)

	time, err := toTimestamp(attrs["data-timestamp"])
	if err != nil {
		fmt.Printf("ERROR: %s\n", err)
	}
	psm.article = &opb.Object{
		Id:        psm.attr("data-story-id"),
		CreatedAt: time,
		Generator: []*opb.Generator{
			{
				Id:   psm.attr("data-author-id"),
				Name: psm.attr("data-author-name"),
			},
		},
	}
}

func (psm *PikabuParserSM) newComment() {
	meta := map[string]string{}
	for _, metaParam := range strings.Split(psm.attr("data-meta"), ";") {
		kv := strings.Split(metaParam, "=")
		if len(kv) == 1 {
			meta[kv[0]] = ""
		} else {
			meta[kv[0]] = kv[1]
		}
	}
	if meta["pid"] == "0" {
		meta["pid"] = psm.article.Id
	}
	t, err := toTimestamp(meta["d"])
	if err != nil {
		// TODO: log time parse error
	}
	psm.comment = append(psm.comment, &opb.Object{
		Id:        psm.attr("data-id"),
		CreatedAt: t,
		Parent:    meta["pid"],
		Generator: []*opb.Generator{
			{
				Id: psm.attr("data-author-id"),
			},
		},
	})
}

func (psm *PikabuParserSM) InDocument() {
	if psm.token.Data == "article" {
		psm.newArticle()
		psm.state = InArticle
	} else if psm.token.Data == "div" && hasClass(psm.token, "comment") {
		psm.newComment()
		psm.state = InComment
	}
}

func (psm *PikabuParserSM) InArticle() {
	if psm.token.Data == "span" && psm.attr("class") == "story__title-link" {
		psm.state = InArticleTitle
	} else if psm.token.Data == "article" {
		psm.state = InDocument
	}
}

func (psm *PikabuParserSM) InArticleTitle() {
	if psm.token.Data == "span" {
		psm.state = InArticle
	} else {
		if psm.article.Content == nil {
			psm.article.Content = []*opb.Content{}
		}
		psm.article.Content = append(psm.article.Content, &opb.Content{
			Text: psm.token.Data,
		})
	}
}

func (psm *PikabuParserSM) InComment() {}

func (psm *PikabuParserSM) InCommentContent() {}

func (psm *PikabuParserSM) Parse() error {
	counter := 0
	for !psm.nextToken() {
		tag := psm.token.Data
		switch psm.state {
		case InDocument:
			psm.InDocument()
		case InArticle:
			psm.InArticle()
		case InArticleTitle:
			psm.InArticleTitle()
		case InComment:
			if tag == "div" && psm.attr("class") == "comment__content" {
				counter = 1
				psm.state = InCommentContent
			}
		case InCommentContent:
			lastComment := psm.comment[len(psm.comment)-1]
			if lastComment.Content == nil {
				lastComment.Content = []*opb.Content{{Mime: "text/html"}}
			}
			if attachment := getAttachment(psm.token); attachment != nil {
				lastComment.Attachment = append(lastComment.Attachment, attachment)
			}
			lastComment.Content[len(lastComment.Content)-1].Text += string(psm.tokenizer.Raw())
			if tag == "div" {
				if psm.token.Type == html.EndTagToken {
					counter--
				} else {
					counter++
				}
			}
			if counter == 0 {
				psm.state = InDocument
			}
		}
	}
	return nil
}

func NewParser(src io.Reader) *PikabuParserSM {
	return &PikabuParserSM{
		tokenizer: html.NewTokenizer(charmap.Windows1251.NewDecoder().Reader(src)),
		state:     InDocument,
		comment:   []*opb.Object{},
	}
}

func main() {
	file, err := os.OpenFile("demo/pikabu_1226249.html", os.O_RDONLY, 0)
	if err != nil {
		os.Exit(-1)
	}
	defer file.Close()

	parser := NewParser(file)
	if err := parser.Parse(); err != nil {
		fmt.Printf("Parse failed: %s\n", err)
	}

	fmt.Printf("Article: %s\n", parser.article)
	for _, c := range parser.comment {
		fmt.Printf("Comment: %s\n", c)
	}
}
