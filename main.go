package main

import (
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"time"

	"golang.org/x/net/html"
	"golang.org/x/text/encoding/charmap"

	"chronicler/parser"
	opb "chronicler/proto"
)

const (
	InError          = -1
	InDocument       = 0
	InArticle        = 1
	InArticleTitle   = 2
	InArticleContent = 3
	InArticleTags    = 4
	InComment        = 5
	InCommentContent = 6
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

func getAttachment(token *html.Token) *opb.Attachment {
	if token.Type != html.StartTagToken && token.Type != html.SelfClosingTagToken {
		return nil
	}
	attrs := attrMap(token)
	urls := []string{}
	for _, attr := range []string{"href", "src", "data-src"} {
		if attrs[attr] != "" {
			urls = append(urls, attrs[attr])
		}
	}
	if len(urls) == 0 {
		return nil
	}
	return &opb.Attachment{Url: urls}
}

type PikabuParserSM struct {
	parser.HtmlParser

	count  int
	state  int
	states map[int]func()

	article *opb.Object
	comment []*opb.Object
}

func (psm *PikabuParserSM) newArticle() {
	token := psm.Token()
	attrs := attrMap(token)

	time, err := toTimestamp(attrs["data-timestamp"])
	if err != nil {
		fmt.Printf("ERROR: %s\n", err)
	}
	psm.article = &opb.Object{
		Id:        psm.Attr("data-story-id"),
		CreatedAt: time,
		Generator: []*opb.Generator{
			{
				Id:   psm.Attr("data-author-id"),
				Name: psm.Attr("data-author-name"),
			},
		},
	}
}

func (psm *PikabuParserSM) newComment() {
	meta := map[string]string{}
	for _, metaParam := range strings.Split(psm.Attr("data-meta"), ";") {
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
		Id:        psm.Attr("data-id"),
		CreatedAt: t,
		Parent:    meta["pid"],
		Generator: []*opb.Generator{
			{
				Id: psm.Attr("data-author-id"),
			},
		},
	})
}

func (psm *PikabuParserSM) InDocument() {
	if psm.Matches("article") {
		psm.newArticle()
		psm.SetState(InArticle)
	} else if psm.Matches("div", "comment") {
		psm.newComment()
		psm.SetState(InComment)
	}
}

func (psm *PikabuParserSM) InArticle() {
	if psm.Matches("span", "story__title-link") {
		psm.SetState(InArticleTitle)
	} else if psm.Matches("div", "story__tags") {
		psm.SetState(InArticleTags)
	} else if psm.Matches("div", "story__content-inner") {
		psm.count = 1
		psm.SetState(InArticleContent)
	} else if psm.Matches("article") {
		psm.SetState(InDocument)
	}
}

func (psm *PikabuParserSM) InArticleTitle() {
	if psm.Matches("span") {
		psm.SetState(InArticle)
	} else {
		psm.article.Content = append(psm.article.Content, &opb.Content{Text: psm.Token().Data})
	}
}

func (psm *PikabuParserSM) InArticleContent() {
	if psm.Matches("div") {
		psm.count++
	} else if psm.Matches("/div") {
		psm.count--
	}
	if psm.count == 0 {
		psm.SetState(InArticle)
		return
	}
	if psm.article.Content == nil {
		psm.article.Content = []*opb.Content{{Mime: "text/html"}}
	}
	psm.article.Content[len(psm.article.Content)-1].Text += psm.Raw()
	if attachment := getAttachment(psm.Token()); attachment != nil {
		psm.article.Attachment = append(psm.article.Attachment, attachment)
	}
}

func (psm *PikabuParserSM) InArticleTags() {
	if psm.Matches("/div") {
		psm.SetState(InArticle)
	} else if psm.Matches("a", "tags__tag") {
		if psm.article.Tag == nil {
			psm.article.Tag = []*opb.Tag{}
		}
		psm.article.Tag = append(psm.article.Tag, &opb.Tag{
			Name: psm.Attr("data-tag"),
			Url:  psm.Attr("href"),
		})
	}
}

func (psm *PikabuParserSM) InComment() {
	if psm.Matches("div", "comment__content") {
		psm.count = 1
		psm.SetState(InCommentContent)
	}
}

func (psm *PikabuParserSM) InCommentContent() {
	if psm.Matches("div") {
		psm.count++
	} else if psm.Matches("/div") {
		psm.count--
	}
	if psm.count == 0 {
		psm.SetState(InDocument)
		return
	}
	lastComment := psm.comment[len(psm.comment)-1]
	if lastComment.Content == nil {
		lastComment.Content = []*opb.Content{{Mime: "text/html"}}
	}
	if attachment := getAttachment(psm.Token()); attachment != nil {
		lastComment.Attachment = append(lastComment.Attachment, attachment)
	}
	lastComment.Content[len(lastComment.Content)-1].Text += string(psm.Raw())
}

func (psm *PikabuParserSM) SetState(state int) error {
	psm.state = state
	return nil
}

func (psm *PikabuParserSM) Parse() error {
	for !psm.NextToken() {
		psm.states[psm.state]()
	}
	return nil
}

func NewParser(src io.Reader) *PikabuParserSM {
	psm := &PikabuParserSM{
		state:   InDocument,
		comment: []*opb.Object{},
	}
	psm.Reader = charmap.Windows1251.NewDecoder().Reader(src)
	psm.states = map[int]func(){
		InDocument:       psm.InDocument,
		InArticle:        psm.InArticle,
		InArticleTitle:   psm.InArticleTitle,
		InArticleContent: psm.InArticleContent,
		InArticleTags:    psm.InArticleTags,
		InComment:        psm.InComment,
		InCommentContent: psm.InCommentContent,
	}
	return psm
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
	// for _, c := range parser.comment {
	// 	fmt.Printf("Comment: %s\n", c)
	// }
}
