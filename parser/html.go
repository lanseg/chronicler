package parser

import (
	"io"
	"slices"
	"strings"

	"golang.org/x/net/html"
)

func attrMap(token *html.Token) map[string]string {
	result := map[string]string{}
	for _, attr := range token.Attr {
		result[attr.Key] = attr.Val
	}
	return result
}

type HtmlParser struct {
	Reader io.Reader

	tokenizer *html.Tokenizer
	token     *html.Token
	attrs     map[string]string
}

func (psm *HtmlParser) NextToken() bool {
	if psm.tokenizer == nil {
		psm.tokenizer = html.NewTokenizer(psm.Reader)
	}
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

func (psm *HtmlParser) HasClass(class string) bool {
	return slices.Contains(strings.Split(psm.Attr("class"), " "), class)
}

func (psm *HtmlParser) Attr(key string) string {
	if len(psm.token.Attr) != 0 && len(psm.attrs) == 0 {
		psm.attrs = attrMap(psm.token)
	}
	if value, ok := psm.attrs[key]; ok {
		return value
	}
	return ""
}

func (psm *HtmlParser) Matches(name string, class ...string) bool {
	t := psm.token.Type
	if t == html.CommentToken || t == html.TextToken || t == html.ErrorToken {
		return false
	}
	if t == html.EndTagToken {
		return name[0] == '/' && len(class) == 0 && name[1:] == psm.Token().Data
	}
	return name == psm.token.Data && (len(class) == 0 || psm.HasClass(class[0]))
}

func (psm *HtmlParser) Token() *html.Token {
	return psm.token
}

func (psm *HtmlParser) Raw() string {
	return string(psm.tokenizer.Raw())
}
