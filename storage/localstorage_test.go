package storage

import (
	"io"
	"reflect"
	"sort"
	"strings"
	"testing"
)

const (
	defaultFile = "test_file"
)

func listResponseUrls(urls ...string) *ListResponse {
	result := &ListResponse{}
	for _, u := range urls {
		result.Items = append(result.Items, StorageItem{
			Url: u,
		})
	}
	return result
}

func write(s Storage, fname string, content []byte) error {
	wc, err := s.Put(&PutRequest{Url: fname})
	if err != nil {
		return err
	}
	defer wc.Close()
	_, err = wc.Write(content)
	if err != nil {
		return err
	}
	return nil
}

func TestLocalStorageMapping(t *testing.T) {
	testdir := t.TempDir()
	testfiles := []string{"file 1", "file 2", "http://file2/something.json?!@#$@!#$@%6"}
	s, err := NewLocalStorage(testdir)
	if err != nil {
		t.Errorf("Cannot initialize storage: %s", err)
	}
	for _, fn := range testfiles {
		if err := write(s, fn, []byte{1, 2, 3}); err != nil {
			t.Errorf("Cannot write to storage: %s", err)
		}
	}
	s, err = NewLocalStorage(testdir)
	if err != nil {
		t.Errorf("Cannot reopen storage: %s", err)
	}

	list, err := s.List(&ListRequest{})
	if err != nil {
		t.Errorf("Error while listing files")
	}
	wantList := listResponseUrls(testfiles...)
	sort.Slice(list.Items, func(a int, b int) bool {
		return list.Items[a].Url < list.Items[b].Url
	})

	if !reflect.DeepEqual(list, wantList) {
		t.Errorf("Expected list to be %q, but got %q", wantList, list)
	}
}

type put struct {
	request *PutRequest
	writes  [][]byte
}

type get struct {
	request   *GetRequest
	wantBytes []byte
	wantErr   string
}

func TestLocalStoragePutFile(t *testing.T) {
	for _, tc := range []struct {
		name       string
		puts       []*put
		gets       []*get
		list       []*ListRequest
		wantList   []*ListResponse
		wantBackup map[string]([]byte)
	}{
		{
			name:     "successful read write",
			puts:     []*put{{writes: [][]byte{[]byte("Hello"), []byte("There")}}},
			gets:     []*get{{wantBytes: []byte("HelloThere")}},
			list:     []*ListRequest{{}},
			wantList: []*ListResponse{listResponseUrls(defaultFile)},
		},
		{
			name:     "empty payload",
			puts:     []*put{{writes: [][]byte{{}, {}, {}}}},
			gets:     []*get{{wantBytes: []byte{}}},
			list:     []*ListRequest{{}},
			wantList: []*ListResponse{listResponseUrls(defaultFile)},
		},
		{
			name:     "zeroes payload",
			puts:     []*put{{writes: [][]byte{{0}, {0}, {0}}}},
			gets:     []*get{{wantBytes: []byte{0, 0, 0}}},
			list:     []*ListRequest{{}},
			wantList: []*ListResponse{listResponseUrls(defaultFile)},
		},
		{
			name:     "bytes payload",
			puts:     []*put{{writes: [][]byte{{1}, {2}, {3}}}},
			gets:     []*get{{wantBytes: []byte{1, 2, 3}}},
			list:     []*ListRequest{{}},
			wantList: []*ListResponse{listResponseUrls(defaultFile)},
		},
		{
			name: "bytes payload",
			puts: []*put{{
				request: &PutRequest{Url: "reтестобщение   +\"*%\"*ç\"*%&"},
				writes:  [][]byte{{1}, {2}, {3}},
			}},
			gets: []*get{{
				request:   &GetRequest{Url: "reтестобщение   +\"*%\"*ç\"*%&"},
				wantBytes: []byte{1, 2, 3},
			}},
			list:     []*ListRequest{{}},
			wantList: []*ListResponse{listResponseUrls("reтестобщение   +\"*%\"*ç\"*%&")},
		},
		{
			name: "multiple puts",
			puts: []*put{
				{request: &PutRequest{Url: "Message"}, writes: [][]byte{{1, 2, 3}}},
				{request: &PutRequest{Url: "Message 2"}, writes: [][]byte{{4, 5, 6}}},
			},
			gets: []*get{
				{request: &GetRequest{Url: "Message"}, wantBytes: []byte{1, 2, 3}},
				{request: &GetRequest{Url: "Message 2"}, wantBytes: []byte{4, 5, 6}},
			},
			list:     []*ListRequest{{}},
			wantList: []*ListResponse{listResponseUrls("Message", "Message 2")},
		},
		{
			name: "save on overwrite",
			puts: []*put{
				{request: &PutRequest{Url: "Message"}, writes: [][]byte{{1, 2, 3}}},
				{request: &PutRequest{Url: "Message", SaveOnOverwrite: true}, writes: [][]byte{{4, 5, 6}}},
				{request: &PutRequest{Url: "Message", SaveOnOverwrite: true}, writes: [][]byte{{7, 8, 9}}},
			},
			gets: []*get{
				{request: &GetRequest{Url: "Message"}, wantBytes: []byte{7, 8, 9}},
			},
			list: []*ListRequest{{WithSnapshots: true}},
			wantList: []*ListResponse{{
				Items: []StorageItem{
					{Url: "Message", Versions: []string{"0000", "0001"}},
				},
			}},
		},
		{
			name: "non existing get",
			gets: []*get{
				{request: &GetRequest{Url: "someurl"}, wantErr: "file does not exist"},
			},
			list:     []*ListRequest{{WithSnapshots: true}},
			wantList: []*ListResponse{},
		},
		{
			name: "list with filter",
			puts: []*put{
				{request: &PutRequest{Url: "filename 1"}, writes: [][]byte{[]byte("file 1")}},
				{request: &PutRequest{Url: "filename 2"}, writes: [][]byte{[]byte("file 2")}},
				{request: &PutRequest{Url: "filename 3"}, writes: [][]byte{[]byte("file 3")}},
			},
			gets: []*get{
				{request: &GetRequest{Url: "filename 1"}, wantBytes: []byte("file 1")},
				{request: &GetRequest{Url: "filename 2"}, wantBytes: []byte("file 2")},
				{request: &GetRequest{Url: "filename 3"}, wantBytes: []byte("file 3")},
			},
			list:     []*ListRequest{{Url: []string{"filename 1", "filename 3"}}},
			wantList: []*ListResponse{listResponseUrls("filename 1", "filename 3")},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			s, err := NewLocalStorage(t.TempDir())
			if err != nil {
				t.Errorf("Cannot create temporary storage: %s", err)
			}
			if tc.puts == nil {
				tc.puts = []*put{}
			}
			for _, put := range tc.puts {
				if put.request == nil {
					put.request = &PutRequest{Url: defaultFile}
				}
				// Write
				wc, err := s.Put(put.request)
				if err != nil {
					t.Errorf("Cannot open writer for the new file %s %s", put.request.Url, err)
				}
				for _, p := range put.writes {
					wc.Write([]byte(p))
				}
				wc.Close()
			}

			// Read
			for _, get := range tc.gets {
				if get.request == nil {
					get.request = &GetRequest{
						Url: defaultFile,
					}
				}
				rc, err := s.Get(get.request)
				if err != nil {
					if !strings.Contains(err.Error(), get.wantErr) {
						t.Errorf("Cannot open reader for the new file %s %s", get.request.Url, err)
					}
					return
				}
				result, err := io.ReadAll(rc)
				if err != nil {
					t.Errorf("Error while reading new file %s %s", get.request.Url, err)
				}

				// Compare
				if !reflect.DeepEqual(result, get.wantBytes) {
					t.Errorf("Expected result and payload to be the same, but got %q and %q",
						result, get.wantBytes)
				}
			}

			for i, lr := range tc.list {
				list, err := s.List(lr)
				if err != nil {
					t.Errorf("Error while listing files")
				}
				sort.Slice(list.Items, func(i, j int) bool {
					return list.Items[i].Url < list.Items[j].Url
				})
				if !reflect.DeepEqual(list, tc.wantList[i]) {
					t.Errorf("Expected list to be %q, but got %q", tc.wantList[i], list)
				}
			}
		})
	}
}
