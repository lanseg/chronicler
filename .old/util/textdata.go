package util

import (
	"regexp"
)

const (
	// https://twitter.com/trusymedvedeva/status/1583893354219065344?s=20&t=Spx5f8ka8yx6VozZE69Lmw
	webLinkRegexp = "((?:https?:)?\\/\\/)?[-a-zA-Z0-9@:%._\\+~#=]{1,256}\\.[a-zA-Z0-9()]{1,6}\\b([-a-zA-Z0-9()@:%_\\+.~#?&//=]*)"
	ytLinkRegexp  = "((?:https?:)?\\/\\/)?((?:www|m)\\.)?((?:youtube\\.com|youtu.be))(\\/(?:[\\w\\-]+\\?v=|embed\\/|v\\/)?)([\\w\\-]+)(\\S+)?"
	ytLinkExact   = "^" + ytLinkRegexp + "$"
	twLinkRegexp  = "^((?:https?:)?\\/\\/)?((?:www|m|mobile)\\.)?twitter.com\\/.*\\/status\\/([0-9]+).*"
)

var (
	ytLink      = regexp.MustCompile(ytLinkExact)
	ytLinkFind  = regexp.MustCompile(ytLinkRegexp)
	webLinkFind = regexp.MustCompile(webLinkRegexp)
	twLink      = regexp.MustCompile(twLinkRegexp)
)

func IsYoutubeLink(link string) bool {
	return ytLink.Match([]byte(link))
}

func FindWebLinks(text string) []string {
	return webLinkFind.FindAllString(text, -1)
}
