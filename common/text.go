package common

import (
	"strings"
	"unicode"
)

func WrapText(text string, maxWidth int) string {
	result := strings.Builder{}
	buffer := strings.Builder{}
	lastLength := 0
	bufferRuneCount := 0
	for _, c := range text {
		if c == '\n' {
			lastLength = 0
			bufferRuneCount = 0
			result.WriteString(buffer.String())
			result.WriteRune('\n')
			buffer.Reset()
			continue
		}
		buffer.WriteRune(c)
		bufferRuneCount += 1
		if unicode.IsSpace(c) || unicode.IsPunct(c) {
			if lastLength+bufferRuneCount >= maxWidth {
				lastLength = bufferRuneCount
				result.WriteRune('\n')
			} else {
				lastLength += bufferRuneCount
			}
			result.WriteString(buffer.String())
			buffer.Reset()
			bufferRuneCount = 0
			continue
		}
	}
	result.WriteString(buffer.String())
	return result.String()
}
