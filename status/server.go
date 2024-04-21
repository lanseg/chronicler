package status

import (
	"io"
	"net"
	"sort"
	"time"

	"google.golang.org/grpc"
	_ "google.golang.org/grpc/encoding/gzip" // Install the gzip compressor
	"google.golang.org/grpc/keepalive"

	col "github.com/lanseg/golang-commons/collections"
	cm "github.com/lanseg/golang-commons/common"

	sp "chronicler/status/status_go_proto"
)

var kaep = keepalive.EnforcementPolicy{
	MinTime:             5 * time.Second, // If a client pings more than once every 5 seconds, terminate the connection
	PermitWithoutStream: true,            // Allow pings even when there are no active streams
}

var kasp = keepalive.ServerParameters{
	MaxConnectionIdle:     15 * time.Second, // If a client is idle for 15 seconds, send a GOAWAY
	MaxConnectionAge:      30 * time.Second, // If any connection is alive for more than 30 seconds, send a GOAWAY
	MaxConnectionAgeGrace: 5 * time.Second,  // Allow 5 seconds for pending RPCs to complete before forcibly closing connections
	Time:                  5 * time.Second,  // Ping the client if it is idle for 5 seconds to ensure the connection is still active
	Timeout:               1 * time.Second,  // Wait 1 second for the ping ack before assuming the connection is dead
}

type statusServer struct {
	sp.UnimplementedStatusServer

	grpcServer *grpc.Server
	address    string
	logger     *cm.Logger

	metrics map[string]*sp.Metric
}

func (s *statusServer) GetStatus(in *sp.GetStatusRequest, out sp.Status_GetStatusServer) error {
	s.logger.Debugf("GetStatusRequest: %v", in)
	result := col.Values(s.metrics)
	sort.Slice(result, func(i int, j int) bool {
		return result[i].Name < result[j].Name
	})
	out.Send(&sp.GetStatusResponse{
		Metric: result,
	})
	return nil
}

func (s *statusServer) PutStatus(out sp.Status_PutStatusServer) error {
	s.logger.Debugf("PutStatusRequest receiver")
	for {
		result, err := out.Recv()
		if err == io.EOF {
			break
		} else if err != nil {
			return err
		}
		s.logger.Debugf("PutStatusRequest: %v", result)
		for _, r := range result.Metric {
			if r.GetValue() == nil {
				delete(s.metrics, r.Name)
			} else {
				s.metrics[r.Name] = r
			}
		}
	}
	return out.SendAndClose(&sp.PutStatusResponse{})
}

func (s *statusServer) Start() error {
	socket, err := net.Listen("tcp", s.address)
	if err != nil {
		return err
	}

	s.grpcServer = grpc.NewServer(
		grpc.KeepaliveEnforcementPolicy(kaep),
		grpc.KeepaliveParams(kasp),
		grpc.MaxSendMsgSize(maxMsgSize),
		grpc.MaxRecvMsgSize(maxMsgSize))
	sp.RegisterStatusServer(s.grpcServer, s)
	s.logger.Infof("Status server listening at %v", socket.Addr())

	go (func() {
		if err := s.grpcServer.Serve(socket); err != nil {
			s.logger.Errorf("Failed to start grpc server: %v", err)
			socket.Close()
		}
	})()
	return nil
}

func (s *statusServer) Stop() {
	s.logger.Infof("Stopping server gracefully")
	s.grpcServer.GracefulStop()
	s.logger.Infof("Server stopped")
}

func NewStatusServer(address string) *statusServer {
	return &statusServer{
		address: address,
		logger:  cm.NewLogger("StatusServer"),
		metrics: map[string]*sp.Metric{},
	}
}
