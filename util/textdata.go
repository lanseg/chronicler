package util

import (
	"regexp"
)

var (
	ytLink = regexp.MustCompile("^((?:https?:)?\\/\\/)?((?:www|m)\\.)?((?:youtube\\.com|youtu.be))(\\/(?:[\\w\\-]+\\?v=|embed\\/|v\\/)?)([\\w\\-]+)(\\S+)?$")
)

func IsYoutubeLink(link string) bool {
	return ytLink.Match([]byte(link))
}
