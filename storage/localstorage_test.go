package storage

import (
	"io"
	"reflect"
	"sort"
	"testing"
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

func TestLocalStoragePutFile(t *testing.T) {
	for _, tc := range []struct {
		name    string
		fname   string
		writes  [][]byte
		wantErr bool
	}{
		{
			name:   "successful read write",
			writes: [][]byte{[]byte("Hello"), []byte("There")},
		},
		{
			name:   "empty payload",
			writes: [][]byte{{}, {}, {}},
		},
		{
			name:   "zeroes payload",
			writes: [][]byte{{0}, {0}, {0}},
		},
		{
			name:   "bytes payload",
			writes: [][]byte{{1}, {2}, {3}},
		},
		{
			name:   "bytes payload",
			fname:  "regergergergergтестовое сообщение   +\"*%\"*ç\"*%&",
			writes: [][]byte{{1}, {2}, {3}},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			s, err := NewLocalStorage(t.TempDir())
			if err != nil {
				t.Errorf("Cannot create temporary storage: %s", err)
			}
			if tc.fname == "" {
				tc.fname = "test_file"
			}
			// Write
			wc, err := s.Put(&PutRequest{Url: tc.fname})
			if err != nil {
				t.Errorf("Cannot open writer for the new file %s %s", tc.fname, err)
			}
			for _, p := range tc.writes {
				wc.Write([]byte(p))
			}
			wc.Close()

			// Read
			rc, err := s.Get(&GetRequest{Url: tc.fname})
			if err != nil {
				t.Errorf("Cannot open reader for the new file %s %s", tc.fname, err)
				return
			}
			result, err := io.ReadAll(rc)
			if err != nil {
				t.Errorf("Error while reading new file %s %s", tc.fname, err)
			}

			// Compare
			want := []byte{}
			for _, p := range tc.writes {
				want = append(want, p...)
			}
			if !reflect.DeepEqual(result, want) {
				t.Errorf("Expected result and payload to be the same, but got %q and %q",
					result, want)
			}

			list, err := s.List(&ListRequest{})
			if err != nil {
				t.Errorf("Error while listing files")
			}
			wantList := &ListResponse{Url: []string{tc.fname}}

			if !reflect.DeepEqual(list, wantList) {
				t.Errorf("Expected list to be %q, but got %q", wantList, list)
			}
		})
	}
}
