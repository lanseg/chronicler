package util

import (
	"fmt"
	"regexp"

	"os/exec"
)

const (
	ytLinkRegex = "^((?:https?:)?\\/\\/)?((?:www|m)\\.)?((?:youtube\\.com|youtu.be))(\\/(?:[\\w\\-]+\\?v=|embed\\/|v\\/)?)([\\w\\-]+)(\\S+)?$"
)

var (
	logger = NewLogger("util")
)

func execute(command string, args []string) error {
	out, err := exec.Command(command, args...).Output()
	if err != nil {
		return err
	}
	output := string(out[:])
	fmt.Println(output)
	return nil
}

func IsYoutubeLink(link string) bool {
	ok, err := regexp.Match(ytLinkRegex, []byte(link))
	if err != nil {
		logger.Warningf("Cannot check if '%s' is youtube link: %s", link, err)
		return false
	}
	return ok
}

func DownloadYoutube(video string, targetDir string) error {
	return execute("yt-dlp", []string{
		"-ciw",
		"-o",
		fmt.Sprintf("\"%s/%%(playlist)s.%%(title)s.%%(ext)s\"", targetDir),
		"-v", video,
	})
}
