package webdriver

import (
	"fmt"
	"os"

	"chronicler/util"

	cm "github.com/lanseg/golang-commons/common"
)

const (
	startupDelay = 10
)

type Firefox struct {
	port          int
	profileFolder string
	logger        *cm.Logger

	Runner *util.Runner
}

func startFirefox(remotePort int, profileFolder string) *Firefox {
	logger := cm.NewLogger("Firefox")
	logger.Infof("Starting firefox on %d and profile %s", remotePort, profileFolder)
	if _, err := os.Stat(profileFolder); os.IsNotExist(err) {
		logger.Infof("Creating profile directory: %s", profileFolder)
		if mkDirErr := os.MkdirAll(profileFolder, 0750); mkDirErr != nil {
			return nil
		}
	}
	runner := util.NewRunner()
	go runner.Execute("firefox", []string{
		"--headless",
		"--marionette",
		"--remote-allow-hosts",
		"127.0.0.1",
		"--profile",
		profileFolder,
		"--remote-debugging-port",
		fmt.Sprintf("%d", remotePort),
	})
	return &Firefox{
		port:          remotePort,
		profileFolder: profileFolder,
		Runner:        runner,
	}
}
