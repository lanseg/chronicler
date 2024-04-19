package status

import (
	"context"
	"io"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/encoding/gzip"
	"google.golang.org/grpc/keepalive"

	cm "github.com/lanseg/golang-commons/common"

	sp "chronicler/status/status_go_proto"
)

const (
	maxMsgSize = 1024 * 1024
)

var kacp = keepalive.ClientParameters{
	Time:                10 * time.Second, // send pings every 10 seconds if there is no activity
	Timeout:             time.Second,      // wait 1 second for ping ack before considering the connection dead
	PermitWithoutStream: true,             // send pings even without active streams
}

func newStatusClient(addr string) (sp.StatusClient, error) {
	conn, err := grpc.Dial(addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithKeepaliveParams(kacp),
		grpc.WithDefaultCallOptions(
			grpc.MaxCallRecvMsgSize(maxMsgSize),
			grpc.UseCompressor(gzip.Name)))
	if err != nil {
		return nil, err
	}
	return sp.NewStatusClient(conn), nil
}

type StatusClient interface {
	PutValue(metric *sp.Metric) error
	GetValues() ([]*sp.Metric, error)
	Start()
	Stop()
}

type noopStatusClient struct {
	StatusClient

	logger *cm.Logger
}

func (nc *noopStatusClient) PutValue(metric *sp.Metric) error {
	nc.logger.Infof("PutValue(%v)", metric)
	return nil
}

func (nc *noopStatusClient) GetValues() ([]*sp.Metric, error) {
	nc.logger.Infof("GetValues()")
	return []*sp.Metric{}, nil
}

func (nc *noopStatusClient) Start() {
	nc.logger.Infof("Start()")
}

func (nc *noopStatusClient) Stop() {
	nc.logger.Infof("Stop()")
}

func NewNoopClient(_ string) (StatusClient, error) {
	return &noopStatusClient{
		logger: cm.NewLogger("NoopStatusClient"),
	}, nil
}

type remoteStatusClient struct {
	StatusClient

	client  sp.StatusClient
	context context.Context
	logger  *cm.Logger

	putter chan *sp.Metric
	done   chan bool
}

func NewStatusClient(addr string) (StatusClient, error) {
	client, err := newStatusClient(addr)
	if err != nil {
		return nil, err
	}
	return &remoteStatusClient{
		client:  client,
		context: context.Background(),
		putter:  make(chan *sp.Metric, 10),
		done:    make(chan bool),
		logger:  cm.NewLogger("RemoteStorage"),
	}, nil
}

func (sc *remoteStatusClient) PutValue(metric *sp.Metric) error {
	sc.putter <- metric
	return nil
}

func (sc *remoteStatusClient) GetValues() ([]*sp.Metric, error) {
	get, err := sc.client.GetStatus(sc.context, &sp.GetStatusRequest{})
	if err != nil {
		return nil, err
	}
	result := []*sp.Metric{}
	for {
		in, err := get.Recv()
		if err == io.EOF {
			break
		} else if err != nil {
			return nil, err
		}
		result = append(result, in.Metric...)
	}
	return result, nil
}

func (sc *remoteStatusClient) Stop() {
	sc.done <- true
}

func (sc *remoteStatusClient) Start() {
	go func() {
		done := false
		sc.logger.Infof("Starting status client")
		put, err := sc.client.PutStatus(sc.context)
		if err != nil {
			sc.logger.Warningf("Could not initialize the connection: %s", err)
			done = true
		}
		for !done {
			select {
			case metric := <-sc.putter:
				if err := put.Send(&sp.PutStatusRequest{
					Metric: []*sp.Metric{metric},
				}); err != nil {
					sc.logger.Warningf("Error while sending the metrics: %s", err)
					continue
				}
			case <-sc.done:
				done = true
			}
		}
		sc.logger.Infof("Shutting down status client")
		_, err = put.CloseAndRecv()
		close(sc.putter)
		close(sc.done)
		sc.logger.Infof("Status client stopped")
	}()
}
