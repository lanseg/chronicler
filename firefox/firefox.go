package firefox

import (
	"fmt"

	"chronicler/util"
)

func StartFirefox(remotePort int, profileFolder string) {
	logger := util.NewLogger("Firefox")
	logger.Infof("Starting firefox on %d and profile %s", remotePort, profileFolder)

	runner := util.NewRunner()
	runner.Execute("firefox", []string{
		"--headless",
		"--marionette",
		"--remote-allow-hosts",
		"127.0.0.1",
		"--profile",
		profileFolder,
		"--remote-debugging-port",
		fmt.Sprintf("%d", remotePort),
	})
}
