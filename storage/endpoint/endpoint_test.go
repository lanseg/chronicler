package endpoint

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"reflect"
	"testing"

	"github.com/lanseg/golang-commons/concurrent"
	"github.com/lanseg/golang-commons/optional"

	rpb "chronicler/records/proto"
	"chronicler/storage"
	ep "chronicler/storage/endpoint_go_proto"
)

const (
	testAddr = "localhost:12345"
)

func newPutRequest(id string, fileId int, name string, data []byte) *ep.PutFileRequest {
	return &ep.PutFileRequest{
		File: &ep.FileDef{
			RecordSetId: id,
			Filename:    name,
		},
		Part: &ep.FilePart{
			FileId: int32(fileId),
			Data: &ep.FilePart_Chunk_{
				Chunk: &ep.FilePart_Chunk{
					ChunkId: int32(0),
					Size:    int32(len(data)),
					Data:    data,
				},
			},
		},
	}
}

type FakeStorage struct {
	storage.Storage

	recordSets []*rpb.RecordSet
	fileData   map[string][]byte
}

func (fs *FakeStorage) SaveRecordSet(r *rpb.RecordSet) error {
	return nil
}

func (fs *FakeStorage) ListRecordSets(_ *rpb.Sorting) optional.Optional[[]*rpb.RecordSet] {
	return optional.Of(fs.recordSets)
}

func (fs *FakeStorage) DeleteRecordSet(id string) error {
	return nil
}

func (fs *FakeStorage) GetFile(id string, filename string) optional.Optional[io.ReadCloser] {
	if data, ok := fs.fileData[fmt.Sprintf("%s_%s", id, filename)]; ok {
		return optional.Of(io.NopCloser(bytes.NewReader(data)))
	}
	return optional.OfError[io.ReadCloser](nil, fmt.Errorf("No file %s/%s", id, filename))
}

func (fs *FakeStorage) PutFile(id string, filename string, reader io.Reader) error {
	data, err := io.ReadAll(reader)
	if err != nil {
		return err
	}
	fs.fileData[fmt.Sprintf("%s_%s", id, filename)] = data
	return nil
}

func genTestFile(length int) []byte {
	result := make([]byte, length)
	for i := 0; i < length; i++ {
		result[i] = byte('a' + i%24)
	}
	return result
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
		return optional.OfError(newEndpointClient(testAddr))
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
		_, err := tb.client.Save(context.Background(), &ep.SaveRequest{
			RecordSet: &rpb.RecordSet{
				Records: []*rpb.Record{
					{},
				},
			},
		})
		if err != nil {
			t.Errorf("Failed to perform Save operation: %s", err)
		}
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

func TestListRequest(t *testing.T) {
	tb, err := setupServer(t)
	if err != nil {
		t.Fatalf("Failed to initialize client-server: %s", err)
	}
	defer tb.tearDown(t)

	for _, tc := range []struct {
		name       string
		recordSets []*rpb.RecordSet
		request    *ep.ListRequest
		want       []*rpb.RecordSet
	}{
		{
			name: "Two record sets in response",
			recordSets: []*rpb.RecordSet{
				{Id: "RecordSet1"},
				{Id: "RecordSet2"},
			},
			request: &ep.ListRequest{},
			want: []*rpb.RecordSet{
				{Id: "RecordSet1"},
				{Id: "RecordSet2"},
			},
		},
		{
			name:       "Empty response",
			recordSets: []*rpb.RecordSet{},
			request:    &ep.ListRequest{},
			want:       []*rpb.RecordSet{},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			tb.storage.recordSets = tc.recordSets
			recv, err := tb.client.List(context.Background(), tc.request)
			if err != nil {
				t.Errorf("Could not perform list request: %s", err)
				return
			}

			result := []*rpb.RecordSet{}
			for {
				rs, err := recv.Recv()
				if err != nil && err != io.EOF {
					t.Errorf("Error while fetching recordset: %s", err)
					return
				}
				if err == io.EOF {
					break
				}
				result = append(result, rs.RecordSet)
			}

			if !reflect.DeepEqual(tc.want, result) {
				t.Errorf("client.List(%s) expected to return %s, but got %s", tc.want, result)
			}
		})
	}
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
				File: []*ep.FileDef{
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
				File: []*ep.FileDef{
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
				File: []*ep.FileDef{
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

func TestPutFile(t *testing.T) {
	tb, err := setupServer(t)
	if err != nil {
		t.Fatalf("Failed to initialize client-server: %s", err)
	}
	defer tb.tearDown(t)

	for _, tc := range []struct {
		name         string
		reqs         []*ep.PutFileRequest
		wantFiles    map[string][]byte
		wantResponse *ep.PutFileResponse
	}{
		{
			name: "put small file",
			reqs: []*ep.PutFileRequest{
				newPutRequest("id0", 0, "fn0", []byte("HELLO")),
				newPutRequest("id0", 1, "fn1", []byte("THERE")),
			},
			wantFiles: map[string][]byte{
				"id0_fn0": []byte("HELLO"),
				"id0_fn1": []byte("THERE"),
			},
			wantResponse: &ep.PutFileResponse{},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			tb.storage.fileData = map[string]([]byte){}
			client, err := tb.client.PutFile(context.Background())
			if err != nil {
				t.Fatalf("Could not open stream: %s", err)
			}

			for _, req := range tc.reqs {
				if err := client.Send(req); err != nil {
					t.Fatalf("Could not send file: %s", err)
				}
			}
			resp, err := client.CloseAndRecv()

			if err != nil || fmt.Sprintf("%s", tc.wantResponse) != fmt.Sprintf("%s", resp) {
				t.Errorf("Expected response to be (%s, nil), but got (%s, %s)", tc.wantResponse, resp, err)
			}
			if !reflect.DeepEqual(tc.wantFiles, tb.storage.fileData) {
				t.Errorf("Expected to get files %s, but got %s", tc.wantFiles, tb.storage.fileData)
			}
		})
	}
}
