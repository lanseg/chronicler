package endpoint

import (
	"context"

	ep "chronicler/storage/endpoint_go_proto"

	cm "github.com/lanseg/golang-commons/common"
)

type storageServer struct {
	ep.UnimplementedStorageServer

	logger *cm.Logger
}

func (s *storageServer) Save(ctx context.Context, in *ep.SaveRequest) (*ep.SaveResponse, error) {
	s.logger.Debugf("Save request: %v", in)
	return &ep.SaveResponse{}, nil
}

func (s *storageServer) List(ctx context.Context, in *ep.ListRequest) (*ep.ListResponse, error) {
	s.logger.Debugf("List request: %v", in)
	return &ep.ListResponse{}, nil
}

func (s *storageServer) Delete(ctx context.Context, in *ep.DeleteRequest) (*ep.DeleteResponse, error) {
	s.logger.Debugf("Delete request: %v", in)
	return &ep.DeleteResponse{}, nil
}

func (s *storageServer) Get(ctx context.Context, in *ep.GetRequest) (*ep.GetResponse, error) {
	s.logger.Debugf("Get request: %v", in)
	return &ep.GetResponse{}, nil
}

func NewStorageServer() *storageServer {
	return &storageServer{
		logger: cm.NewLogger("Storage"),
	}
}
