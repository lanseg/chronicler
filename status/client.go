package status

import (
	sp "chronicler/status/status_go_proto"
	"context"
	"io"

	cm "github.com/lanseg/golang-commons/common"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/encoding/gzip"
)

const (
	maxMsgSize = 1024 * 1024
)

func newStatusClient(addr string) (sp.StatusClient, error) {
	conn, err := grpc.Dial(addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
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
		for !done {
			select {
			case metric := <-sc.putter:
				put, err := sc.client.PutStatus(sc.context)
				if err != nil {
					sc.logger.Warningf("Could not initialize the connection: %s", err)
					continue
				}
				if err := put.Send(&sp.PutStatusRequest{
					Metric: []*sp.Metric{metric},
				}); err != nil {
					sc.logger.Warningf("Error while sending the metrics: %s", err)
					continue
				}

				_, err = put.CloseAndRecv()

			case <-sc.done:
				sc.logger.Infof("Shutting down status client")
				close(sc.putter)
				close(sc.done)
				done = true
			}
		}
		sc.logger.Infof("Status client stopped")
	}()
}
