package endpoint

import (
	"context"
	"fmt"
	"reflect"
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

func genTestFile(length int) []byte {
	result := make([]byte, length)
	for i := 0; i < length; i++ {
		result[i] = byte('a' + i%24)
	}
	return result
}

type FakeStorage struct {
	storage.Storage

	fileData map[string][]byte
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
	if data, ok := fs.fileData[fmt.Sprintf("%s_%s", id, filename)]; ok {
		return optional.Of(data)
	}
	return optional.OfError[[]byte](nil, fmt.Errorf("No file %s/%s", id, filename))
}

type testBed struct {
	storage  *FakeStorage
	server   *storageServer
	client   ep.StorageClient
	tearDown func(tb testing.TB)
}

func setupServer(tb testing.TB) (*testBed, error) {
	storage := &FakeStorage{
		fileData: map[string][]byte{},
	}
	server := NewStorageServer(testAddr, storage)
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
		client:  client,
		server:  server,
		storage: storage,
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
}

func TestGetFile(t *testing.T) {
	tb, err := setupServer(t)
	if err != nil {
		t.Fatalf("Failed to initialize client-server: %s", err)
	}
	defer tb.tearDown(t)

	for _, tc := range []struct {
		name        string
		storageData map[string][]byte
		request     *ep.GetFileRequest
		want        []*FileData
	}{
		{
			name: "One name different records gives different files",
			storageData: map[string][]byte{
				"123_SomeFile": []byte("Hello world"),
				"456_SomeFile": []byte("Hello there"),
			},
			request: &ep.GetFileRequest{
				File: []*ep.GetFileRequest_FileDef{
					{RecordSetId: "123", Filename: "SomeFile"},
					{RecordSetId: "456", Filename: "SomeFile"},
				},
			},
			want: []*FileData{
				{Data: []byte("Hello world")},
				{Data: []byte("Hello there")},
			},
		},
		{
			name: "FileData for missing file is not an error",
			storageData: map[string][]byte{
				"123_SomeFile": []byte("Hello world"),
			},
			request: &ep.GetFileRequest{
				File: []*ep.GetFileRequest_FileDef{
					{RecordSetId: "123", Filename: "SomeFile"},
					{RecordSetId: "456", Filename: "NonExistingFile"},
				},
			},
			want: []*FileData{
				{Data: []byte("Hello world")},
				{Error: "No file 456/NonExistingFile"},
			},
		},
		{
			name: "File bigger than one chunk",
			storageData: map[string][]byte{
				"123_SomeFile": genTestFile(chunkSize + 10),
			},
			request: &ep.GetFileRequest{
				File: []*ep.GetFileRequest_FileDef{
					{RecordSetId: "123", Filename: "SomeFile"},
				},
			},
			want: []*FileData{
				{Data: genTestFile(chunkSize + 10)},
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			tb.storage.fileData = tc.storageData
			result, err := ReadAll(tb.client.GetFile(context.Background(), tc.request))
			if err != nil {
				t.Errorf("Expected GetFile(%v) = %v, but got error %s", tc.request, tc.want, err)
				return
			}

			if !reflect.DeepEqual(tc.want, result) {
				t.Errorf("Expected GetFile(%v) = %v, but got %v", tc.request, tc.want, result)
			}
		})
	}
}
