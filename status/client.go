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

type StatusClient struct {
	client  sp.StatusClient
	context context.Context
	logger  *cm.Logger

	putter chan *sp.Metric
	done   chan bool
}

func NewStatusClient(addr string) (*StatusClient, error) {
	client, err := newStatusClient(addr)
	if err != nil {
		return nil, err
	}
	return &StatusClient{
		client:  client,
		context: context.Background(),
		putter:  make(chan *sp.Metric, 10),
		done:    make(chan bool),
		logger:  cm.NewLogger("RemoteStorage"),
	}, nil
}

func (sc *StatusClient) PutValue(metric *sp.Metric) {
	sc.putter <- metric
}

func (sc *StatusClient) GetValues() ([]*sp.Metric, error) {
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

func (sc *StatusClient) Stop() {
	sc.done <- true
}

func (sc *StatusClient) Start() {
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
