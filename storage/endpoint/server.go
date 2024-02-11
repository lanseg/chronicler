package endpoint

import (
	"context"
	"net"

	"google.golang.org/grpc"

	rpb "chronicler/records/proto"
	"chronicler/storage"
	ep "chronicler/storage/endpoint_go_proto"

	cm "github.com/lanseg/golang-commons/common"
)

type storageServer struct {
	ep.UnimplementedStorageServer

	grpcServer  *grpc.Server
	address     string
	baseStorage storage.Storage
	logger      *cm.Logger
}

func (s *storageServer) Save(ctx context.Context, in *ep.SaveRequest) (*ep.SaveResponse, error) {
	s.logger.Debugf("Save request: %v", in)
	return &ep.SaveResponse{}, nil
}

func (s *storageServer) List(ctx context.Context, in *ep.ListRequest) (*ep.ListResponse, error) {
	s.logger.Debugf("List request: %v", in)
	return &ep.ListResponse{
		RecordSets: s.baseStorage.ListRecordSets().OrElse([]*rpb.RecordSet{}),
	}, nil
}

func (s *storageServer) Delete(ctx context.Context, in *ep.DeleteRequest) (*ep.DeleteResponse, error) {
	s.logger.Debugf("Delete request: %v", in)
	for _, id := range in.RecordSetIds {
		s.baseStorage.DeleteRecordSet(id)
	}
	return &ep.DeleteResponse{}, nil
}

func (s *storageServer) Get(ctx context.Context, in *ep.GetRequest) (*ep.GetResponse, error) {
	s.logger.Debugf("Get request: %v", in)
	sets := []*rpb.RecordSet{}
	for _, id := range in.RecordSetIds {
		s.logger.Debugf("Result: %s", id)
	}
	return &ep.GetResponse{
		RecordSets: sets,
	}, nil
}

func (s *storageServer) Start() error {
	socket, err := net.Listen("tcp", s.address)
	if err != nil {
		return err
	}

	s.grpcServer = grpc.NewServer()
	ep.RegisterStorageServer(s.grpcServer, s)
	s.logger.Infof("Storage server listening at %v", socket.Addr())

	go (func() {
		if err := s.grpcServer.Serve(socket); err != nil {
			s.logger.Errorf("Failed to start grpc server: %v", err)
			socket.Close()
		}
	})()
	return nil
}

func (s *storageServer) Stop() {
	s.logger.Infof("Stopping server gracefully")
	s.grpcServer.GracefulStop()
	s.logger.Infof("Server stopped")
}

func NewStorageServer(address string, str storage.Storage) *storageServer {
	return &storageServer{
		logger:      cm.NewLogger("Storage"),
		baseStorage: str,
		address:     address,
	}
}
