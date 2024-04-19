package main

import (
	"fmt"
	"os"
	"sync"

	cm "github.com/lanseg/golang-commons/common"

	"chronicler/status"
)

var (
	logger = cm.NewLogger("storageserver")
)

type Config struct {
	StatusServerPort *int `json:"statusServerPort"`
}

func main() {
	cfg := cm.OrExit(cm.GetConfig[Config](os.Args[1:], "config"))
	if cfg.StatusServerPort == nil {
		logger.Warningf("No status server port configured ")
		os.Exit(-1)
	}
	logger.Infof("Config.StatusServerPort: %d", *cfg.StatusServerPort)
	srv := status.NewStatusServer(fmt.Sprintf("localhost:%d", *cfg.StatusServerPort))
	srv.Start()
	wg := sync.WaitGroup{}
	wg.Add(1)
	wg.Wait()
}
