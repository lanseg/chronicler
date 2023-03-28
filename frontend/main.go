package main

import (
	"flag"
	"os"

	"chronicler/frontend"
	"chronicler/util"
)

var (
	port        = flag.Int("port", 8080, "Port for the http server")
	storageRoot = flag.String("storage_root", "chronicler_storage", "A local folder to save downloads.")
	staticRoot  = flag.String("static_root", "./static", "Root directory with the web page files. ")

	log = util.NewLogger("main")
)

func main() {
	flag.Parse()
	logger := util.NewLogger("main")

	cwd, _ := os.Getwd()
	logger.Infof("Currect directory: %s", cwd)
	logger.Infof("Storage root: %s", *storageRoot)
	logger.Infof("Static files root: %s", *storageRoot)
	logger.Infof("Starting server on port %d", *port)

	server := frontend.NewServer(*port, *storageRoot, *staticRoot)
	server.Start()
}
