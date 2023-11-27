package tinyhtml

import (
	"strings"
	"unicode"
)

type State int

const (
	ERROR                   = State(0)
	DATA                    = State(1)
	TAG_OPEN                = State(2)
	TAG_NAME                = State(3)
	TAG_PARAM_BEFORE        = State(4)
	TAG_PARAM               = State(5)
	TAG_PARAM_AFTER         = State(6)
	TAG_VALUE_BEFORE        = State(7)
	TAG_VALUE               = State(8)
	TAG_VALUE_QUOTED_SINGLE = State(9)
	TAG_VALUE_QUOTED_DOUBLE = State(10)
	TAG_CLOSING             = State(11)
	TAG_SELF_CLOSING        = State(12)
	SCRIPT_DATA             = State(13)
	MAYBE_TAG_COMMENT       = State(14)
	TAG_COMMENT             = State(15)
)

var (
	names = []string{
		"ERROR", "DATA", "TAG_OPEN",
		"TAG_NAME", "TAG_PARAM_BEFORE",
		"TAG_PARAM", "TAG_PARAM_AFTER",
		"TAG_VALUE_BEFORE", "TAG_VALUE",
		"TAG_VALUE_QUOTED_SINGLE",
		"TAG_VALUE_QUOTED_DOUBLE",
		"TAG_CLOSING", "TAG_SELF_CLOSING",
		"SCRIPT_DATA", "MAYBE_TAG_COMMENT", "TAG_COMMENT",
	}
)

func (s State) String() string {
	return names[s]
}

type Token struct {
	value     string
	tokenType State
	start     int
	end       int
}

// Tokenizer
type Tokenizer struct {
	buffer strings.Builder

	tokens     []*Token
	lastOfType map[State]*Token
}

func (t *Tokenizer) push(pos int, tokenType State) {
	start := 0
	if len(t.tokens) > 0 {
		start = t.tokens[len(t.tokens)-1].end
	}
	t.pushToken(&Token{
		value:     t.buffer.String(),
		tokenType: tokenType,
		start:     start,
		end:       pos + 1,
	})
}

func (t *Tokenizer) pushToken(tok *Token) {
	t.tokens = append(t.tokens, tok)
	t.lastOfType[tok.tokenType] = t.tokens[len(t.tokens)-1]
	t.buffer.Reset()
}

func (t *Tokenizer) feed(r rune) {
	t.buffer.WriteRune(r)
}

func tokenize(doc string) []*Token {
	state := DATA
	tokens := &Tokenizer{
		lastOfType: map[State]*Token{},
		tokens:     []*Token{},
	}
    asRunes := []rune(doc)
	for i, r := range asRunes {
		switch state {
		case ERROR:
		case DATA:
			if r == '<' {
				state = TAG_OPEN
				tokens.push(i, DATA)
			} else if tok, ok := tokens.lastOfType[TAG_NAME]; ok && tok.value == "script" {
				state = SCRIPT_DATA
				tokens.feed(r)
			} else {
				tokens.feed(r)
			}
		case SCRIPT_DATA:
			datavalue := tokens.buffer.String()
			if strings.HasSuffix(datavalue, "</script>") {
				tokens.push(i, SCRIPT_DATA)
				tokens.pushToken(&Token{
					value:     "/script",
					tokenType: TAG_NAME,
					start:     i - 9,
					end:       i,
				})
				if r == '<' {
					state = TAG_OPEN
				} else {
					state = DATA
					tokens.feed(r)
				}
			} else {
				tokens.feed(r)
			}
		case TAG_OPEN:
			if r == '>' {
				state = ERROR
			} else if r == '!' {
				state = MAYBE_TAG_COMMENT
				tokens.feed(r)
			} else if !unicode.IsSpace(r) {
				tokens.feed(r)
				state = TAG_NAME
			}
		case MAYBE_TAG_COMMENT:
			if r == '-' {
				state = TAG_COMMENT
				tokens.feed(r)
			} else {
				tokens.feed(r)
				state = TAG_NAME
			}
		case TAG_COMMENT:
			datavalue := tokens.buffer.String()
			if r == '>' {
				if strings.HasSuffix(datavalue, "--") {
					tokens.push(i, TAG_COMMENT)
					state = DATA
				}
			} else {
				tokens.feed(r)
			}
		case TAG_NAME:
			if r == '>' {
				tokens.push(i, TAG_NAME)
				state = DATA
			} else if !unicode.IsSpace(r) {
				tokens.feed(r)
			} else {
				tokens.push(i, TAG_NAME)
				state = TAG_PARAM_BEFORE
			}
		case TAG_PARAM_BEFORE:
			if r == '>' {
				state = DATA
			} else if !unicode.IsSpace(r) {
				tokens.feed(r)
				state = TAG_PARAM
			}
		case TAG_PARAM:
			if r == '>' {
				tokens.push(i, TAG_PARAM)
				state = DATA
			} else if r == '=' {
				tokens.push(i, TAG_PARAM)
				state = TAG_VALUE_BEFORE
			} else if unicode.IsSpace(r) {
				tokens.push(i, TAG_PARAM)
				state = TAG_PARAM_AFTER
			} else {
				tokens.feed(r)
			}
		case TAG_PARAM_AFTER:
			if r == '>' {
				state = DATA
			} else if r == '=' {
				state = TAG_VALUE_BEFORE
			}
		case TAG_VALUE_BEFORE:
			if r == '>' {
				state = DATA
			} else if r == '"' {
				state = TAG_VALUE_QUOTED_DOUBLE
			} else if r == '\'' {
				state = TAG_VALUE_QUOTED_SINGLE
			} else if unicode.IsSpace(r) {
				state = TAG_VALUE
			}
		case TAG_VALUE:
			if r == '>' {
				tokens.push(i, TAG_VALUE)
				state = DATA
			} else if unicode.IsSpace(r) {
				tokens.push(i, TAG_VALUE)
				state = TAG_PARAM_BEFORE
			} else {
				tokens.feed(r)
			}
		case TAG_VALUE_QUOTED_SINGLE:
			if r == '\'' {
				tokens.push(i, TAG_VALUE)
				state = TAG_PARAM_BEFORE
			} else {
				tokens.feed(r)
			}
		case TAG_VALUE_QUOTED_DOUBLE:
			if r == '"' {
				tokens.push(i, TAG_VALUE)
				state = TAG_PARAM_BEFORE
			} else {
				tokens.feed(r)
			}
		}
	}
    tokens.push(len(asRunes), DATA)
	return tokens.tokens
}
