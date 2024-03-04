package endpoint

import (
	"context"
	"io"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/encoding/gzip"

	rpb "chronicler/records/proto"
	"chronicler/storage"
	ep "chronicler/storage/endpoint_go_proto"

	cm "github.com/lanseg/golang-commons/common"
	"github.com/lanseg/golang-commons/optional"
)

func newEndpointClient(addr string) (ep.StorageClient, error) {
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

func NewRemoteStorage(addr string) (storage.Storage, error) {
	client, err := newEndpointClient(addr)
	if err != nil {
		return nil, err
	}
	return &remoteStorage{
		client: client,

		logger: cm.NewLogger("RemoteStorage"),
	}, nil
}

type remoteStorage struct {
	storage.Storage

	context context.Context
	client  ep.StorageClient
	logger  *cm.Logger
}

func (rs *remoteStorage) SaveRecordSet(r *rpb.RecordSet) error {
	_, err := rs.client.Save(rs.context, &ep.SaveRequest{})
	return err
}

func (rs *remoteStorage) ListRecordSets() optional.Optional[[]*rpb.RecordSet] {
	return optional.Of([]*rpb.RecordSet{})
}

func (rs *remoteStorage) GetRecordSet(id string) optional.Optional[*rpb.RecordSet] {
	return optional.Of[*rpb.RecordSet](nil)
}

func (rs *remoteStorage) DeleteRecordSet(id string) error {
	return nil
}

func (rs *remoteStorage) GetFile(id string, filename string) optional.Optional[io.ReadCloser] {
	return optional.Of[io.ReadCloser](nil)
}

func (rs *remoteStorage) PutFile(id string, filename string, src io.Reader) error {
	return nil
}
