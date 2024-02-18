package endpoint

import (
	"context"
	"fmt"
	"testing"

	rpb "chronicler/records/proto"
	"chronicler/storage"
	ep "chronicler/storage/endpoint_go_proto"

	"github.com/lanseg/golang-commons/concurrent"
	"github.com/lanseg/golang-commons/optional"
)

const (
	testAddr = "localhost:12345"
)

type FakeStorage struct {
	storage.Storage
}

func (fs *FakeStorage) SaveRecordSet(r *rpb.RecordSet) error {
	return nil
}

func (fs *FakeStorage) ListRecordSets() optional.Optional[[]*rpb.RecordSet] {
	return optional.Of([]*rpb.RecordSet{})
}

func (fs *FakeStorage) DeleteRecordSet(id string) error {
	return nil
}

func (fs *FakeStorage) GetFile(id string, filename string) optional.Optional[[]byte] {
	return optional.Of([]byte{})
}

type testBed struct {
	server   *storageServer
	client   ep.StorageClient
	tearDown func(tb testing.TB)
}

func setupServer(tb testing.TB) (*testBed, error) {
	server := NewStorageServer(testAddr, &FakeStorage{})
	if err := server.Start(); err != nil {
		return nil, err
	}

	client, err := concurrent.WaitForSomething(func() optional.Optional[ep.StorageClient] {
		return optional.OfError(NewEndpointClient(testAddr))
	}).Get()
	if err != nil {
		return nil, err
	}

	return &testBed{
		client: client,
		server: server,
		tearDown: func(tb testing.TB) {
			server.Stop()
		},
	}, nil
}

func TestStorage(t *testing.T) {
	tb, err := setupServer(t)
	if err != nil {
		t.Fatalf("Failed to initialize client-server: %s", err)
	}
	defer tb.tearDown(t)

	t.Run("Save", func(t *testing.T) {
		result, err := tb.client.Save(context.Background(), &ep.SaveRequest{})
		if err != nil {
			t.Errorf("Failed to perform Save operation: %s", err)
		}
		fmt.Println(result)
	})

	t.Run("List", func(t *testing.T) {
		result, err := tb.client.List(context.Background(), &ep.ListRequest{})
		if err != nil {
			t.Errorf("Failed to perform List operation: %s", err)
		}
		fmt.Println(result)
	})

	t.Run("Delete", func(t *testing.T) {
		result, err := tb.client.Delete(context.Background(), &ep.DeleteRequest{})
		if err != nil {
			t.Errorf("Failed to perform Delete operation: %s", err)
		}
		fmt.Println(result)
	})

	t.Run("Get", func(t *testing.T) {
		result, err := tb.client.Get(context.Background(), &ep.GetRequest{})
		if err != nil {
			t.Errorf("Failed to perform Get operation: %s", err)
		}
		fmt.Println(result)
	})

	t.Run("GetFile", func(t *testing.T) {
		result, err := tb.client.GetFile(context.Background(), &ep.GetFileRequest{})
		if err != nil {
			t.Errorf("Failed to perform GetFile operation: %s", err)
		}
		fmt.Println(result)
	})

}
