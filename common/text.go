package common

import (
	"fmt"
	"hash/fnv"
	"mime"
	"net/url"
	"path/filepath"
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

func hash(input string) string {
	hash := fnv.New32()
	hash.Write([]byte(input))
	return fmt.Sprintf("_%x", hash.Sum([]byte{}))
}

func Sanitize(input string, maxLength int) string {
	builder := strings.Builder{}
	hashLength := 9
	for i, r := range input {
		if i >= maxLength-hashLength && maxLength > 0 {
			builder.WriteString(hash(input))
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
	return builder.String()
}

func SanitizeWithExt(remotePath string, mimeType string, maxLength int) string {
	ext := ""
	maybeUrl, err := url.Parse(remotePath)
	if err == nil {
		path := maybeUrl.Path
		pathEnd := strings.Index(path, "?")
		if pathEnd != -1 {
			path = path[:pathEnd]
		}
		ext = filepath.Ext(path)
	} else {
		ext = filepath.Ext(remotePath)
	}
	if ext == "" && mimeType != "" {
		exts, err := mime.ExtensionsByType(mimeType)
		if err == nil && len(exts) > 0 {
			ext = exts[0]
		}
	}
	safeUrl := Sanitize(remotePath, maxLength)
	if ext == "" || len(safeUrl) == len(remotePath) {
		return safeUrl
	}
	safeUrl = safeUrl[:len(safeUrl)-len(ext)-1]
	if safeUrl[len(safeUrl)-1] != '.' {
		safeUrl += "."
	}
	return safeUrl + ext
}
