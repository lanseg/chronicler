package util

import (
	"regexp"
)

const (
	webLinkRegexp = "((?:https?:)?\\/\\/)?[-a-zA-Z0-9@:%._\\+~#=]{1,256}\\.[a-zA-Z0-9()]{1,6}\\b([-a-zA-Z0-9()@:%_\\+.~#?&//=]*)"
	ytLinkRegexp  = "((?:https?:)?\\/\\/)?((?:www|m)\\.)?((?:youtube\\.com|youtu.be))(\\/(?:[\\w\\-]+\\?v=|embed\\/|v\\/)?)([\\w\\-]+)(\\S+)?"
	ytLinkExact   = "^" + ytLinkRegexp + "$"
)

var (
	ytLink      = regexp.MustCompile(ytLinkExact)
	ytLinkFind  = regexp.MustCompile(ytLinkRegexp)
	webLinkFind = regexp.MustCompile(webLinkRegexp)
)

func IsYoutubeLink(link string) bool {
	return ytLink.Match([]byte(link))
}

func FindYoutubeLinks(text string) []string {
	return ytLinkFind.FindAllString(text, -1)
}

func FindWebLinks(text string) []string {
	return webLinkFind.FindAllString(text, -1)
}
