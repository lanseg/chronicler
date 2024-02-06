package main

import (
	"fmt"
	"net"
	"os"

	"google.golang.org/grpc"

	"chronicler/storage/endpoint"
	ep "chronicler/storage/endpoint_go_proto"

	cm "github.com/lanseg/golang-commons/common"
)

var (
	logger = cm.NewLogger("main")
)

type Config struct {
	Port *int    `json:"storageServerPort"`
	Root *string `json:"storageRoot"`
}

func main() {
	cfg, err := cm.GetConfig[Config](os.Args[1:], "config")
	if err != nil {
		logger.Errorf("Could not load config: %v", err)
		os.Exit(-1)
	}

	socket, err := net.Listen("tcp", fmt.Sprintf("localhost:%d", *cfg.Port))
	if err != nil {
		logger.Errorf("Failed to create server socket: %v", err)
	}

	grpcServer := grpc.NewServer()
	ep.RegisterStorageServer(grpcServer, endpoint.NewStorageServer())
	logger.Infof("Storage server listening at %v", socket.Addr())

	if err := grpcServer.Serve(socket); err != nil {
		logger.Errorf("Could not start gRPC server: %v", err)
		os.Exit(-1)
	}
}
