package util

import (
	"fmt"
	"io"
	"os/exec"
	"path/filepath"

	cm "github.com/lanseg/golang-commons/common"
)

type Runner struct {
	logger *cm.Logger

	done chan error
}

func toStdout(pipe io.ReadCloser) error {
	buf := make([]byte, 1024)
	for {
		n, err := pipe.Read(buf)
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}
		fmt.Print(string(buf[:n]))
	}
}

func (r *Runner) Execute(command string, args []string) {
	r.logger.Debugf("Start %s %s", command, args)
	cmd := exec.Command(command, args...)
	errPipe, err := cmd.StderrPipe()
	if err != nil {
		r.done <- err
		return
	}
	outPipe, err := cmd.StdoutPipe()
	if err != nil {
		r.done <- err
		return
	}
	if err := cmd.Start(); err != nil {
		r.done <- err
		return
	}

	go toStdout(outPipe)
	go toStdout(errPipe)

	if err := cmd.Wait(); err != nil {
		r.done <- err
		return
	}
	r.logger.Debugf("End %s %s", command, args)
	r.done <- nil
}

func DownloadYoutube(video string, targetDir string) error {
	r := NewRunner()
	go r.Execute("yt-dlp", []string{
		"-ciw",
		"-o",
		filepath.Join(targetDir, "%(playlist)s.%(title)s.%(ext)s"),
		"-v", video,
	})
	return <-r.done
}

func NewRunner() *Runner {
	return &Runner{
		logger: cm.NewLogger("Runner"),
		done:   make(chan error),
	}
}
