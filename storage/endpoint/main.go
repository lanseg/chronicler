package main

import (
	"fmt"
	"os"
	"sync"

	"chronicler/downloader"
	"chronicler/storage"
	"chronicler/storage/endpoint"
	"chronicler/webdriver"

	cm "github.com/lanseg/golang-commons/common"
)

var (
	logger = cm.NewLogger("main")
)

type Config struct {
	StorageServerPort *int    `json:"storageServerPort"`
	StorageRoot       *string `json:"storageRoot"`
}

func main() {
	cfg, err := cm.GetConfig[Config](os.Args[1:], "config")
	if err != nil {
		logger.Errorf("Could not load config: %v", err)
		os.Exit(-1)
	}

	logger.Infof("Config.StorageRoot: %s", *cfg.StorageRoot)
	logger.Infof("Config.StorageServerPort: %s", *cfg.StorageServerPort)

	stg := storage.NewStorage(*cfg.StorageRoot, webdriver.NewFakeBrowser(nil), downloader.NewNoopDownloader())
	eps := endpoint.NewStorageServer(fmt.Sprintf("localhost:%d", *cfg.StorageServerPort), stg)
	eps.Start()
	wg := sync.WaitGroup{}
	wg.Add(1)
	wg.Wait()
}
