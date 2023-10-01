package firefox

import (
	"fmt"
	"os"
	"time"

	"chronicler/util"

	"github.com/lanseg/golang-commons/optional"
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
	Driver WebDriver
}

func NewFirefox(profileFolder string, remotePort int) optional.Optional[*Firefox] {
	optional.OfErrorNullable[Firefox](nil, os.MkdirAll(profileFolder, 0750))
	return optional.OfNullable(&Firefox{})
}

func StartFirefox(remotePort int, profileFolder string) *Firefox {
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
	logger.Infof("Waiting for %ds before connecting to firefox", startupDelay)
	time.Sleep(startupDelay * time.Second)

	driver, _ := ConnectMarionette("127.0.0.1", remotePort).Get()

	return &Firefox{
		port:          remotePort,
		profileFolder: profileFolder,
		Runner:        runner,
		Driver:        driver,
	}
}
