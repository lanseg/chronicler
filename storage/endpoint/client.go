package endpoint

import (
	"bytes"
	"context"
	"fmt"
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
		client:  client,
		context: context.Background(),
		logger:  cm.NewLogger("RemoteStorage"),
	}, nil
}

type remoteStorage struct {
	storage.Storage

	context context.Context
	client  ep.StorageClient
	logger  *cm.Logger
}

func (rs *remoteStorage) SaveRecordSet(r *rpb.RecordSet) error {
	_, err := rs.client.Save(rs.context, &ep.SaveRequest{
		RecordSet: r,
	})
	return err
}

func (rs *remoteStorage) ListRecordSets() optional.Optional[[]*rpb.RecordSet] {
	recv, err := rs.client.List(rs.context, &ep.ListRequest{})
	if err != nil {
		return optional.OfError[[]*rpb.RecordSet](nil, err)
	}
	result := []*rpb.RecordSet{}
	for {
		rs, err := recv.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			return optional.OfError[[]*rpb.RecordSet](nil, err)
		}
		result = append(result, rs.RecordSet)
	}
	return optional.Of(result)
}

func (rs *remoteStorage) GetRecordSet(id string) optional.Optional[*rpb.RecordSet] {
	return optional.MapErr(optional.OfError(rs.client.Get(rs.context, &ep.GetRequest{
		RecordSetIds: []string{id},
	})), func(resp *ep.GetResponse) (*rpb.RecordSet, error) {
		rss := resp.RecordSets
		if len(rss) != 1 {
			return nil, fmt.Errorf("Expected to get 1 record set, but got %d", len(rss))
		}
		return rss[0], nil
	})
}

func (rs *remoteStorage) DeleteRecordSet(id string) error {
	_, err := rs.client.Delete(rs.context, &ep.DeleteRequest{RecordSetIds: []string{id}})
	return err
}

func (rs *remoteStorage) GetFile(id string, filename string) optional.Optional[io.ReadCloser] {
	files, err := ReadAll(rs.client.GetFile(rs.context, &ep.GetFileRequest{
		File: []*ep.FileDef{
			{RecordSetId: id, Filename: filename},
		},
	}))
	if err != nil {
		return optional.OfError[io.ReadCloser](nil, err)
	}
	if files == nil || len(files) == 0 || files[0] == nil {
		return optional.OfError[io.ReadCloser](nil, fmt.Errorf("File %s/%s not found", id, filename))
	}
	return optional.Of[io.ReadCloser](io.NopCloser(bytes.NewReader(files[0].Data)))
}

func (rs *remoteStorage) PutFile(id string, filename string, src io.Reader) error {
	rs.logger.Infof("Saving file to %s/%s", id, filename)
	put, err := rs.client.PutFile(rs.context)
	if err != nil {
		return err
	}
	buf := make([]byte, chunkSize)
	for chunkId := 0; ; chunkId++ {
		size, err := src.Read(buf)
		if err == io.EOF && size == 0 {
			break
		}
		rs.logger.Infof("Read %d bytes from %s/%s", size, id, filename)
		part := &ep.PutFileRequest{
			Part: &ep.FilePart{
				FileId: int32(0),
			},
		}
		if chunkId == 0 {
			part.File = &ep.FileDef{
				RecordSetId: id,
				Filename:    filename,
			}
		}

		if size > 0 {
			part.Part.Data = &ep.FilePart_Chunk_{
				Chunk: &ep.FilePart_Chunk{
					ChunkId: int32(chunkId),
					Data:    buf[:size],
				},
			}

		} else if err != nil {
			part.Part.Data = &ep.FilePart_Error_{
				Error: &ep.FilePart_Error{
					Error: err.Error(),
				},
			}
		}
		if err := put.Send(part); err != nil {
			rs.logger.Infof("Error while sending file part %s\n", part.File)
			break
		}
	}
	_, err = put.CloseAndRecv()
	return err
}
