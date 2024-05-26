package endpoint

import (
	"bytes"
	"context"
	"fmt"
	"net"

	cm "github.com/lanseg/golang-commons/common"
	"google.golang.org/grpc"
	_ "google.golang.org/grpc/encoding/gzip" // Install the gzip compressor
	"google.golang.org/grpc/health"
	healthgrpc "google.golang.org/grpc/health/grpc_health_v1"
	healthpb "google.golang.org/grpc/health/grpc_health_v1" // Separating proto for readability

	// Separating proto for readability
	rpb "chronicler/records/proto"
	"chronicler/storage"
	ep "chronicler/storage/endpoint_go_proto"
)

const (
	chunkSize  = 1024 * 1024
	maxMsgSize = 16 * 1024 * 1024
)

type storageServer struct {
	ep.UnimplementedStorageServer

	grpcServer   *grpc.Server
	healthServer *health.Server

	address     string
	baseStorage storage.Storage
	logger      *cm.Logger
}

func (s *storageServer) Save(ctx context.Context, in *ep.SaveRequest) (*ep.SaveResponse, error) {
	if in.RecordSet == nil || len(in.RecordSet.Records) == 0 {
		return nil, fmt.Errorf("Empty save request")
	}
	s.logger.Debugf("Save request, recordset of size %v", len(in.RecordSet.Records))
	s.baseStorage.SaveRecordSet(in.RecordSet)
	return &ep.SaveResponse{}, nil
}

func (s *storageServer) List(in *ep.ListRequest, out ep.Storage_ListServer) error {
	s.logger.Debugf("List request: %v", in)
	// TODO: Return errors properly
	s.baseStorage.ListRecordSets(in.Request).IfPresent(func(rss []*rpb.RecordSet) {
		for i, rs := range rss {
			if err := out.Send(&ep.ListResponse{RecordSet: rs}); err != nil {
				break
			}
			s.logger.Debugf("Sent %d of %d recordsets\n", i, len(rss))
		}
	})
	return nil
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
		s.baseStorage.GetRecordSet(id).IfPresent(func(rs *rpb.RecordSet) {
			sets = append(sets, rs)
		})
	}
	return &ep.GetResponse{
		RecordSets: sets,
	}, nil
}

func (s *storageServer) GetFile(in *ep.GetFileRequest, out ep.Storage_GetFileServer) error {
	s.logger.Debugf("Get file request: %v", in)
	for i, file := range in.File {
		f, err := s.baseStorage.GetFile(file.RecordSetId, file.Filename).Get()
		if err != nil {
			s.logger.Warningf("Could not read file #%d (%s): %s", i, f, err)
			out.Send(&ep.GetFileResponse{
				Part: &ep.FilePart{
					FileId: int32(i),
					Data: &ep.FilePart_Error_{
						Error: &ep.FilePart_Error{
							Error: err.Error(),
						},
					},
				},
			})
			continue
		}
		if err = WriteAll(out, f, i, chunkSize); err != nil {
			s.logger.Warningf("Could not send file %d/chunk %d: %s", i, 0, err)
			continue
		}
		s.logger.Debugf("Written file #%d (%s/%s)", i, file.RecordSetId, file.Filename)
	}
	return nil
}

func (s *storageServer) PutFile(out ep.Storage_PutFileServer) error {
	defs := map[int32]*ep.FileDef{}
	data := map[int32][]byte{}
	for {
		f, err := out.Recv()
		if err != nil {
			break
		}
		if f.File != nil && defs[f.Part.FileId] == nil {
			s.logger.Infof("Put file %s", f.File)
			defs[f.Part.FileId] = f.File
			data[f.Part.FileId] = []byte{}
		}
		// TODO: Handle errors here
		if f.Part.GetChunk() != nil {
			data[f.Part.FileId] = append(data[f.Part.FileId], f.Part.GetChunk().Data...)
		} else if f.Part.GetError() != nil {
			s.logger.Warningf("A client-side read error for %q: %s, got partial content.",
				defs[f.Part.FileId], f.Part.GetError())
			break
		}
	}
	for index, dataBytes := range data {
		def := defs[index]
		s.baseStorage.PutFile(def.RecordSetId, def.Filename, bytes.NewReader(dataBytes))
	}
	return out.SendAndClose(&ep.PutFileResponse{})
}

func (s *storageServer) Start() error {
	socket, err := net.Listen("tcp", s.address)
	if err != nil {
		return err
	}

	s.healthServer = health.NewServer()
	s.healthServer.SetServingStatus("", healthpb.HealthCheckResponse_SERVING)

	s.grpcServer = grpc.NewServer(
		grpc.MaxSendMsgSize(maxMsgSize),
		grpc.MaxRecvMsgSize(maxMsgSize))
	ep.RegisterStorageServer(s.grpcServer, s)
	healthgrpc.RegisterHealthServer(s.grpcServer, s.healthServer)

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
	if s.grpcServer != nil {
		s.grpcServer.GracefulStop()
		s.logger.Infof("Server stopped")
	} else {
		s.logger.Infof("Server already stopped")
	}
}

func NewStorageServer(address string, str storage.Storage) *storageServer {
	return &storageServer{
		logger:      cm.NewLogger("Storage"),
		baseStorage: str,
		address:     address,
	}
}
