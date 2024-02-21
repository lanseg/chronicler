package main

import (
	"fmt"
	"os"

	"chronicler/frontend"
	sep "chronicler/storage/endpoint"

	cm "github.com/lanseg/golang-commons/common"
)

type FrontendConfig struct {
	StaticRoot        *string `json:staticRoot`
	FrontendPort      *int    `json:frontendPort`
	StorageServerPort *int    `json:"storageServerPort"`
}

func main() {
	logger := cm.NewLogger("main")
	cfg, err := cm.GetConfig[FrontendConfig](os.Args[1:], "config")
	if err != nil {
		logger.Errorf("Could not load configs: %s", err)
		os.Exit(-1)
	}

	logger.Infof("StaticRoot: %s", *cfg.StaticRoot)
	logger.Infof("FrontendPort: %d", *cfg.FrontendPort)
	logger.Infof("StorageServerPort: %d", *cfg.StorageServerPort)

	storageClient, err := sep.NewEndpointClient(fmt.Sprintf("localhost:%d", *cfg.StorageServerPort))
	if err != nil {
		logger.Errorf("Could not connect to the storage: %v", err)
		os.Exit(-1)
	}
	server := frontend.NewServer(*cfg.FrontendPort, *cfg.StaticRoot, storageClient)
	if err := server.ListenAndServe(); err != nil {
		logger.Errorf("Failed to start server: %s", err)
	}
}
