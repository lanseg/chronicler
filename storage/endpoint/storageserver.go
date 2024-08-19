package main

import (
	"fmt"
	"os"
	"sync"

	cm "github.com/lanseg/golang-commons/common"

	"chronicler/storage"
	"chronicler/storage/endpoint"
)

var (
	logger = cm.NewLogger("storageserver")
)

type Config struct {
	StorageServerPort *int    `json:"storageServerPort"`
	StorageRoot       *string `json:"storageRoot"`
}

func main() {
	cfg := cm.OrExit(cm.GetConfig[Config](os.Args[1:], "config"))
	logger.Infof("Config.StorageRoot: %s", *cfg.StorageRoot)
	logger.Infof("Config.StorageServerPort: %s", *cfg.StorageServerPort)

	stg := storage.NewLocalStorage(*cfg.StorageRoot)
	eps := endpoint.NewStorageServer(fmt.Sprintf("0.0.0.0:%d", *cfg.StorageServerPort), stg)
	eps.Start()
	wg := sync.WaitGroup{}
	wg.Add(1)
	wg.Wait()
}
