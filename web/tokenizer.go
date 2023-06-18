package tokenizer

import (
	"strings"
	"unicode"

	"github.com/lanseg/golang-commons/collections"
)

func isSpace(s string) bool {
	for _, c := range s {
		if !unicode.IsSpace(c) {
			return false
		}
	}
	return true
}

type Token struct {
	Name   string
	Text   string
	Params []collections.Pair[string, string]
}

type Tokenizer struct {
	pos    int
	text   []rune
	tokens []*Token

	parser func()
}

func (p *Tokenizer) charAt() rune {
	return p.text[p.pos]
}

func (p *Tokenizer) isChar(r rune) bool {
	return !p.isEnd() && r == p.text[p.pos]
}

func (p *Tokenizer) isSpace() bool {
	return unicode.IsSpace(p.charAt())
}

func (p *Tokenizer) isEnd() bool {
	return p.pos >= len(p.text)
}

func (p *Tokenizer) next() {
	if p.isEnd() {
		return
	}
	p.pos++
}

func (p *Tokenizer) skipSpace() {
	for !p.isEnd() && p.isSpace() {
		p.next()
	}
}

func (p *Tokenizer) lastToken() *Token {
	if len(p.tokens) == 0 {
		return nil
	}
	return p.tokens[len(p.tokens)-1]
}

func (p *Tokenizer) addTextToken(content string) *Token {
	p.tokens = append(p.tokens, &Token{
		Text: content,
	})
	return p.lastToken()
}

func (p *Tokenizer) addTagToken(name string, params []collections.Pair[string, string]) *Token {
	p.tokens = append(p.tokens, &Token{
		Name:   name,
		Params: params,
	})
	return p.lastToken()
}

func (p *Tokenizer) parseParamName() string {
	paramName := strings.Builder{}
	for !p.isEnd() && !p.isChar('>') && !p.isChar('=') && !p.isSpace() {
		paramName.WriteRune(p.charAt())
		p.next()
	}
	return paramName.String()
}

func (p *Tokenizer) parseParamValue() string {
	paramValue := strings.Builder{}
	char := p.charAt()
	quoted := char == '"' || char == '\''
	if quoted {
		p.next()
	}
	for !p.isEnd() && !p.isChar('>') && !((quoted && p.isChar(char)) || (!quoted && p.isSpace())) {
		paramValue.WriteRune(p.charAt())
		p.next()
		for p.isChar('\\') {
			paramValue.WriteRune(p.charAt())
			p.next()
			paramValue.WriteRune(p.charAt())
			p.next()
		}
	}
	if quoted {
		p.next()
	}
	return paramValue.String()
}

func (p *Tokenizer) parseParamList() {
	params := []collections.Pair[string, string]{}
	for !p.isEnd() && !p.isChar('>') {
		paramName := p.parseParamName()
		p.skipSpace()
		if p.pos >= len(p.text) || p.isChar('>') {
			break
		}
		if p.isChar('=') {
			p.next()
			paramValue := p.parseParamValue()
			params = append(params, collections.AsPair(paramName, paramValue))
		} else {
			params = append(params, collections.AsPair(paramName, ""))
		}
		p.skipSpace()
	}
	p.tokens[len(p.tokens)-1].Params = params
	p.parser = p.parseDocument
	p.next()
}

func (p *Tokenizer) parseTag() {
	tokenBuffer := strings.Builder{}
	for !p.isEnd() && !p.isSpace() && !p.isChar('>') {
		tokenBuffer.WriteRune(p.charAt())
		p.next()
	}
	if tokenBuffer.Len() != 0 {
		p.addTagToken(tokenBuffer.String(), []collections.Pair[string, string]{})
	}
	p.skipSpace()
	if p.isChar('>') {
		p.parser = p.parseDocument
		p.next()
	} else {
		p.parser = p.parseParamList
	}
}

func (p *Tokenizer) parseContent() {
	tokenBuffer := strings.Builder{}
	for !p.isEnd() && !p.isChar('<') {
		tokenBuffer.WriteRune(p.charAt())
		p.next()
	}
	result := tokenBuffer.String()
	if len(result) > 0 && !isSpace(result) {
		p.addTextToken(result)
	}
	if !p.isEnd() && p.isChar('<') {
		p.parser = p.parseTag
		p.next()
	}
}

func (p *Tokenizer) parseScriptContent() {
	tokenBuffer := strings.Builder{}
	for !p.isEnd() && !strings.HasSuffix(tokenBuffer.String(), "</script>") {
		tokenBuffer.WriteRune(p.charAt())
		p.next()
	}
	result := tokenBuffer.String()
	if strings.HasSuffix(result, "</script>") {
		p.pos -= len("</script>")
		result = result[:len(result)-len("</script>")]
		p.parser = p.parseContent
	} else {
		p.parser = p.parseDocument
	}
	if len(result) != 0 && !isSpace(result) {
		p.addTextToken(result)
	}
}

func (p *Tokenizer) parseDocument() {
	switch p.lastToken().Name {
	case "script":
		p.parser = p.parseScriptContent
	default:
		p.parser = p.parseContent
	}
}

func (p *Tokenizer) tokenize(data string) {
	p.text = []rune(data)
	p.parser = p.parseDocument

	for !p.isEnd() {
		p.parser()
	}
}

func Tokenize(input string) []*Token {
	tok := &Tokenizer{
		tokens: []*Token{
			{
				Name: "root",
			},
		},
	}
	tok.tokenize(input)
	return tok.tokens[1:]
}
