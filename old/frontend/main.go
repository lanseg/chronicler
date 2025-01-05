package main

import (
	"os"

	cm "github.com/lanseg/golang-commons/common"

	"chronicler/frontend"
	"chronicler/status"
	sep "chronicler/storage/endpoint"
)

type FrontendConfig struct {
	StaticRoot    *string `json:"staticRoot"`
	FrontendPort  *int    `json:"frontendPort"`
	StatusServer  *string `json:"statusServer"`
	StorageServer *string `json:"storageServer"`
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
	logger.Infof("Config.StatusServer: %s", *cfg.StatusServer)
	logger.Infof("StorageServer: %d", *cfg.StorageServer)

	stats := cm.OrExit(status.NewStatusClient(*cfg.StatusServer))
	stats.Start()
	storage, err := sep.NewRemoteStorage(*cfg.StorageServer)

	if err != nil {
		logger.Errorf("Could not connect to the storage: %v", err)
		os.Exit(-1)
	}

	server := frontend.NewServer(*cfg.FrontendPort, *cfg.StaticRoot, storage, stats)
	if err := server.ListenAndServe(); err != nil {
		logger.Errorf("Failed to start server: %s", err)
	}
}
