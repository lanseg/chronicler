package main

import (
	"flag"
	"os"

	"chronicler/frontend"

	cm "github.com/lanseg/golang-commons/common"
)

var (
	port        = flag.Int("port", 8080, "Port for the http server")
	storageRoot = flag.String("storage_root", "chronicler_storage", "A local folder to save downloads.")
	staticRoot  = flag.String("static_root", "frontend/static", "Root directory with the web page files. ")
)

func main() {
	flag.Parse()
	logger := cm.NewLogger("main")

	cwd, _ := os.Getwd()
	logger.Infof("Currect directory: %s", cwd)
	logger.Infof("Storage root: %s", *storageRoot)
	logger.Infof("Static files root: %s", *staticRoot)
	logger.Infof("Starting server on port %d", *port)

	server := frontend.NewServer(*port, *storageRoot, *staticRoot)
	if err := server.ListenAndServe(); err != nil {
		logger.Errorf("Failed to start server: %s", err)
	}
}
