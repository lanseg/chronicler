package util

import (
	"fmt"

	"os/exec"
	"path/filepath"
)

func execute(command string, args []string) error {
	out, err := exec.Command(command, args...).Output()
	if err != nil {
		return err
	}
	fmt.Println(string(out))
	return nil
}

func DownloadYoutube(video string, targetDir string) error {
	return execute("yt-dlp", []string{
		"-ciw",
		"-o",
		fmt.Sprintf("\"%s\"", filepath.Join(targetDir, "%(playlist)s.%(title)s.%(ext)s")),
		"-v", video,
	})
}
