package storage

import (
	"io"
	"reflect"
	"sort"
	"testing"
)

const (
	defaultFile = "test_file"
)

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
	wantList := &ListResponse{Url: testfiles}
	sort.Strings(list.Url)

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
}

func TestLocalStoragePutFile(t *testing.T) {
	for _, tc := range []struct {
		name       string
		puts       []*put
		gets       []*get
		list       []*ListRequest
		wantList   []*ListResponse
		wantBackup map[string]([]byte)
		wantErr    bool
	}{
		{
			name:     "successful read write",
			puts:     []*put{{writes: [][]byte{[]byte("Hello"), []byte("There")}}},
			gets:     []*get{{wantBytes: []byte("HelloThere")}},
			list:     []*ListRequest{{}},
			wantList: []*ListResponse{{Url: []string{defaultFile}}},
		},
		{
			name:     "empty payload",
			puts:     []*put{{writes: [][]byte{{}, {}, {}}}},
			gets:     []*get{{wantBytes: []byte{}}},
			list:     []*ListRequest{{}},
			wantList: []*ListResponse{{Url: []string{defaultFile}}},
		},
		{
			name:     "zeroes payload",
			puts:     []*put{{writes: [][]byte{{0}, {0}, {0}}}},
			gets:     []*get{{wantBytes: []byte{0, 0, 0}}},
			list:     []*ListRequest{{}},
			wantList: []*ListResponse{{Url: []string{defaultFile}}},
		},
		{
			name:     "bytes payload",
			puts:     []*put{{writes: [][]byte{{1}, {2}, {3}}}},
			gets:     []*get{{wantBytes: []byte{1, 2, 3}}},
			list:     []*ListRequest{{}},
			wantList: []*ListResponse{{Url: []string{defaultFile}}},
		},
		{
			name: "bytes payload",
			puts: []*put{{
				request: &PutRequest{Url: "regergergergergтестовое сообщение   +\"*%\"*ç\"*%&"},
				writes:  [][]byte{{1}, {2}, {3}},
			}},
			gets: []*get{{
				request:   &GetRequest{Url: "regergergergergтестовое сообщение   +\"*%\"*ç\"*%&"},
				wantBytes: []byte{1, 2, 3},
			}},
			list:     []*ListRequest{{}},
			wantList: []*ListResponse{{Url: []string{"regergergergergтестовое сообщение   +\"*%\"*ç\"*%&"}}},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			s, err := NewLocalStorage(t.TempDir())
			if err != nil {
				t.Errorf("Cannot create temporary storage: %s", err)
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
					t.Errorf("Cannot open reader for the new file %s %s", get.request.Url, err)
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
				if !reflect.DeepEqual(list, tc.wantList[i]) {
					t.Errorf("Expected list to be %q, but got %q", tc.wantList[i], list)
				}
			}
		})
	}
}
