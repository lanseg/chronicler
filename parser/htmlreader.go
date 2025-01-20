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

type HtmlReader interface {
	NextToken() bool
	HasClass(class string) bool
	Attr(key string) string
	Matches(name string, class ...string) bool
	Raw() string
}

type htmlReader struct {
	HtmlReader

	reader io.Reader

	tokenizer *html.Tokenizer
	token     *html.Token
	attrs     map[string]string
}

func (hr *htmlReader) NextToken() bool {
	if hr.tokenizer == nil {
		hr.tokenizer = html.NewTokenizer(hr.reader)
	}
	if tokenType := hr.tokenizer.Next(); tokenType == html.ErrorToken {
		return true
	}
	token := hr.tokenizer.Token()
	hr.token = &token
	if hr.attrs == nil || len(hr.attrs) != 0 {
		hr.attrs = map[string]string{}
	}
	return false
}

func (hr *htmlReader) HasClass(class string) bool {
	return slices.Contains(strings.Split(hr.Attr("class"), " "), class)
}

func (hr *htmlReader) Attr(key string) string {
	if len(hr.token.Attr) != 0 && len(hr.attrs) == 0 {
		hr.attrs = attrMap(hr.token)
	}
	if value, ok := hr.attrs[key]; ok {
		return value
	}
	return ""
}

func (hr *htmlReader) Matches(name string, class ...string) bool {
	t := hr.token.Type
	if t == html.CommentToken || t == html.TextToken || t == html.ErrorToken {
		return false
	}
	if t == html.EndTagToken {
		return name[0] == '/' && len(class) == 0 && name[1:] == hr.token.Data
	}
	return name == hr.token.Data && (len(class) == 0 || hr.HasClass(class[0]))
}

func (hr *htmlReader) Raw() string {
	return string(hr.tokenizer.Raw())
}

func NewHtmlReader(reader io.Reader) HtmlReader {
	return &htmlReader{
		reader: reader,
	}
}
