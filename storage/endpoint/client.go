package endpoint

import (
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/encoding/gzip"

	ep "chronicler/storage/endpoint_go_proto"
)

func NewEndpointClient(addr string) (ep.StorageClient, error) {
	conn, err := grpc.Dial(addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithDefaultCallOptions(
			grpc.MaxCallRecvMsgSize(maxMsgSize),
			grpc.UseCompressor(gzip.Name)))
	if err != nil {
		return nil, err
	}
	return ep.NewStorageClient(conn), nil
}
