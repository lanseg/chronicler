package tokenizer

import (
	"fmt"
	"strings"
	"unicode"
)

func isSpace(s string) bool {
	for _, c := range s {
		if !unicode.IsSpace(c) {
			return false
		}
	}
	return true
}

type Param struct {
	Key   string
	Value string
}

func (p Param) String() string {
	return fmt.Sprintf("{\"%s\"=\"%s\"}", p.Key, p.Value)
}

type Token struct {
	Name   string
	Text   string
	Params []Param
}

type Tokenizer struct {
	pos          int
	text         string
	tokens       []*Token
	parseHistory string

	parser func()
}

func (p *Tokenizer) charAt() rune {
	return rune(p.text[p.pos])
}

func (p *Tokenizer) isChar(r rune) bool {
	return !p.isEnd() && r == rune(p.text[p.pos])
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
	p.parseHistory += string(p.charAt())
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

func (p *Tokenizer) addTagToken(name string, params []Param) *Token {
	p.tokens = append(p.tokens, &Token{
		Name:   name,
		Params: params,
	})
	return p.lastToken()
}

func (p *Tokenizer) parseParamName() string {
	paramName := ""
	for !p.isEnd() && !p.isChar('>') && !p.isChar('=') && !p.isSpace() {
		paramName += string(p.charAt())
		p.next()
	}
	return paramName
}

func (p *Tokenizer) parseParamValue() string {
	paramValue := ""
	char := p.charAt()
	quoted := char == '"' || char == '\''
	if quoted {
		p.next()
	}
	for !p.isEnd() && !p.isChar('>') && !((quoted && p.isChar(char)) || (!quoted && p.isSpace())) {
		paramValue += string(p.charAt())
		p.next()
		for p.isChar('\\') {
			paramValue += string(p.charAt())
			p.next()
			paramValue += string(p.charAt())
			p.next()
		}
	}
	if quoted {
		p.next()
	}
	return paramValue
}

func (p *Tokenizer) parseParamList() {
	params := []Param{}
	for !p.isEnd() && !p.isChar('>') {
		paramName := p.parseParamName()
		p.skipSpace()
		if p.pos >= len(p.text) || p.isChar('>') {
			break
		}
		if p.isChar('=') {
			p.next()
			paramValue := p.parseParamValue()
			params = append(params, Param{Key: paramName, Value: paramValue})
		} else {
			params = append(params, Param{Key: paramName})
		}
		p.skipSpace()
	}
	p.tokens[len(p.tokens)-1].Params = params
	p.parser = p.parseDocument
	p.next()
}

func (p *Tokenizer) parseTag() {
	tokenBuffer := ""
	for !p.isEnd() && !p.isSpace() && !p.isChar('>') {
		tokenBuffer += string(p.charAt())
		p.next()
	}
	if len(tokenBuffer) != 0 {
		p.addTagToken(tokenBuffer, []Param{})
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
	tokenBuffer := ""
	for !p.isEnd() && !p.isChar('<') {
		tokenBuffer += string(p.charAt())
		p.next()
	}
	if len(tokenBuffer) != 0 && !isSpace(tokenBuffer) {
		p.addTextToken(tokenBuffer)
	}
	if !p.isEnd() && p.isChar('<') {
		p.parser = p.parseTag
		p.next()
	}
}

func (p *Tokenizer) parseScriptContent() {
	tokenBuffer := ""
	for !p.isEnd() && !strings.HasSuffix(tokenBuffer, "</script>") {
		tokenBuffer += string(p.charAt())
		p.next()
	}
	if len(tokenBuffer) != 0 && !isSpace(tokenBuffer) {
		p.addTextToken(tokenBuffer)
	}
	p.parser = p.parseDocument
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
	p.text = data
	p.parser = p.parseDocument

	for p.pos < len(data) {
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
