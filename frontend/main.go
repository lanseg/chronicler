package main

import (
	"os"

	"chronicler/frontend"

	cm "github.com/lanseg/golang-commons/common"
)

type FrontendConfig struct {
    StorageRoot *string `json:storageRoot`
    StaticRoot *string `json:staticRoot`
    FrontendPort *int `json:frontendPort`
}

func main() {
    logger := cm.NewLogger("main")
    cfg, err := cm.GetConfig[FrontendConfig](os.Args[1:], "config")
    if err != nil {
      logger.Errorf("Could not load configs: %s", err)
      os.Exit(-1)
    }

	logger.Infof("Storage root: %s", *cfg.StorageRoot)
	logger.Infof("Static files root: %s", *cfg.StaticRoot)
	logger.Infof("Starting server on port %d", *cfg.FrontendPort)

	server := frontend.NewServer(*cfg.FrontendPort, *cfg.StorageRoot, *cfg.StaticRoot)
	if err := server.ListenAndServe(); err != nil {
		logger.Errorf("Failed to start server: %s", err)
	}
}
