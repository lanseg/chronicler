package main

import (
	"context"
	"fmt"
	"net"
	"os"

	"google.golang.org/grpc"

	ep "chronicler/storage/endpoint"

	cm "github.com/lanseg/golang-commons/common"
)

var (
	logger = cm.NewLogger("main")
)

type Config struct {
	Port *int    `json:"storageServerPort"`
	Root *string `json:"storageRoot"`
}

type storageServer struct {
	ep.UnimplementedStorageServer
}

func (s *storageServer) Save(ctx context.Context, in *ep.SaveRequest) (*ep.SaveResponse, error) {
	return nil, nil
}

func (s *storageServer) List(ctx context.Context, in *ep.ListRequest) (*ep.ListResponse, error) {
	return nil, nil
}

func (s *storageServer) Delete(ctx context.Context, in *ep.DeleteRequest) (*ep.DeleteResponse, error) {
	return nil, nil
}

func (s *storageServer) Get(ctx context.Context, in *ep.GetRequest) (*ep.GetResponse, error) {
	return nil, nil
}

func newStorageServer() *storageServer {
	return &storageServer{}
}

func main() {
	cfg, err := cm.GetConfig[Config](os.Args[1:], "config")
	if err != nil {
		logger.Errorf("Could not load config: %v", err)
		os.Exit(-1)
	}

	socket, err := net.Listen("tcp", fmt.Sprintf(":%d", *cfg.Port))
	if err != nil {
		logger.Errorf("Failed to create server socket: %v", err)
	}

	grpcServer := grpc.NewServer()
	ep.RegisterStorageServer(grpcServer, newStorageServer())
	logger.Infof("Storage server listening at %v", socket.Addr())

	if err := grpcServer.Serve(socket); err != nil {
		logger.Errorf("Could not start gRPC server: %v", err)
		os.Exit(-1)
	}
}
