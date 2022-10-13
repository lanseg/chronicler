package util

import (
	"fmt"

	"os/exec"
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

func DownloadYoutube(video string, targetDir string) error {
	return execute("yt-dlp", []string{
		"-ciw",
		"-o",
		fmt.Sprintf("\"%s/%%(playlist)s.%%(title)s.%%(ext)s\"", targetDir),
		"-v", video,
	})
}
