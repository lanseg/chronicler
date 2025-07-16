package common

import (
	"fmt"
	"hash/fnv"
	"mime"
	"net/url"
	"path/filepath"
	"strings"
	"unicode"
	"unicode/utf8"
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

func hash(input string) string {
	hash := fnv.New32()
	hash.Write([]byte(input))
	return fmt.Sprintf("_%x", hash.Sum([]byte{}))
}

func Sanitize(input string, maxLength int) string {
	if input == "" {
		return ""
	}
	builder := strings.Builder{}
	hashLength := 8
	for i, r := range input {
		if maxLength > 0 && i >= maxLength-hashLength-1 {
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
	builder.WriteString(hash(input))
	return builder.String()
}

func SanitizeWithExt(input string, mimeType string, maxLength int) string {
	if input == "" {
		return ""
	}
	ext := ""
	maybeUrl, err := url.Parse(input)
	if err == nil {
		path := maybeUrl.Path
		pathEnd := strings.Index(path, "?")
		if pathEnd != -1 {
			path = path[:pathEnd]
		}
		ext = filepath.Ext(path)
	} else {
		ext = filepath.Ext(input)
	}
	if ext == "" && mimeType != "" {
		exts, err := mime.ExtensionsByType(mimeType)
		if err == nil && len(exts) > 0 {
			ext = exts[0]
		}
	}
	if ext == "" || maxLength == 0 {
		return Sanitize(input, maxLength) + ext
	}
	return Sanitize(input, min(utf8.RuneCountInString(input), maxLength)-len(ext)) + ext
}
