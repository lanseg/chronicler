package common

import (
	"fmt"
	"hash/fnv"
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

func SanitizeUrl(remotePath string, maxLength int) string {
	builder := strings.Builder{}
	addHash := false
	for i, r := range remotePath {
		if maxLength > 0 && i >= maxLength-9 {
			addHash = true
			break
		}
		if (r >= 'a' && r <= 'z') ||
			(r >= 'A' && r <= 'Z') ||
			(r >= '0' && r <= '9') ||
			r == '-' || r == '_' || r == '.' {
			builder.WriteRune(r)
		} else {
			builder.WriteRune('_')
		}
	}
	if addHash {
		hash := fnv.New32()
		hash.Write([]byte(remotePath))
		builder.WriteString(fmt.Sprintf("_%x", hash.Sum([]byte{})))
	}
	return builder.String()
}
