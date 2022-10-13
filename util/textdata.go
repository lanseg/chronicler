package util

import (
    "regexp"
)

const (
    ytLinkRegex = "^((?:https?:)?\\/\\/)?((?:www|m)\\.)?((?:youtube\\.com|youtu.be))(\\/(?:[\\w\\-]+\\?v=|embed\\/|v\\/)?)([\\w\\-]+)(\\S+)?$"
)

func IsYoutubeLink(link string) bool {
	ok, err := regexp.Match(ytLinkRegex, []byte(link))
	if err != nil {
		logger.Warningf("Cannot check if '%s' is youtube link: %s", link, err)
		return false
	}
	return ok
}
