package storage

import (
	"bytes"
	"fmt"
	"io"
	"reflect"
	"testing"

	cm "github.com/lanseg/golang-commons/common"
	"github.com/lanseg/golang-commons/optional"

	rpb "chronicler/records/proto"
)

func newRecordSet(id int, name string) *rpb.RecordSet {
	rs := &rpb.RecordSet{
		Id: fmt.Sprintf("%s %d", name, id),
	}
	return rs
}

func newRecordSetFull(id int, name string, nrecords int) *rpb.RecordSet {
	set := newRecordSet(id, name)
	set.Records = []*rpb.Record{}
	for i := 0; i < nrecords; i++ {
		set.Records = append(set.Records, &rpb.Record{
			Source:      &rpb.Source{},
			TextContent: fmt.Sprintf("Record %d", i),
			Time:        10000000,
		})
	}
	return set
}

func TestDeleteRecord(t *testing.T) {
	t.Run("Create delete record", func(t *testing.T) {
		s := NewLocalStorage(t.TempDir())

		recordSets := []*rpb.RecordSet{}
		for i := 1; i < 10; i++ {
			rs := newRecordSetFull(i, "what", 10)
			rs.Id = cm.UUID4()
			if saveError := s.SaveRecordSet(rs); saveError != nil {
				t.Errorf("Error while saving a record set: %s", saveError)
			}
			recordSets = append(recordSets, rs)
		}

		query := &rpb.Query{Sorting: &rpb.Sorting{Field: rpb.Sorting_FETCH_TIME}}
		fromStorageBefore, _ := s.ListRecordSets(query).Get()
		if removeErr := s.DeleteRecordSet(recordSets[2].Id); removeErr != nil {
			t.Errorf("Error while removing a record set: %s", removeErr)
		}

		fromStorageAfter, _ := s.ListRecordSets(query).Get()
		if len(fromStorageAfter) != len(fromStorageBefore)-1 {
			t.Errorf("Record was not removed. Before: %d, After: %d",
				len(fromStorageBefore), len(fromStorageAfter))
		}
	})

}

type FileDef struct {
	rsId string
	name string
	data []byte
}

func TestPutFile(t *testing.T) {
	rsId1 := cm.UUID4()
	for _, tc := range []struct {
		name  string
		toPut []*FileDef
	}{
		{
			name: "Put single file",
			toPut: []*FileDef{
				{rsId1, "filename", []byte("Hello there")},
			},
		}, {
			name: "Put multiple files same id different names",
			toPut: []*FileDef{
				{rsId1, "filename", []byte("Hello there")},
				{rsId1, "filename2", []byte("Hello world")},
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			s := NewLocalStorage(t.TempDir())
			if saveError := s.SaveRecordSet(&rpb.RecordSet{Id: rsId1}); saveError != nil {
				t.Errorf("Error while saving a request: %s", saveError)
			}

			for _, fd := range tc.toPut {
				if err := s.PutFile(fd.rsId, fd.name, bytes.NewReader(fd.data)); err != nil {
					t.Errorf("Error while putting file  %s/%s: %s", fd.rsId, fd.name, err)
					return
				}

				got, err := optional.MapErr(s.GetFile(fd.rsId, fd.name),
					func(rc io.ReadCloser) ([]byte, error) {
						defer rc.Close()
						return io.ReadAll(rc)
					}).Get()

				if err != nil || !reflect.DeepEqual(fd.data, got) {
					t.Errorf("Expected GetFile(%s, %s) to return (%s, nil), but got (%s, %s)",
						fd.rsId, fd.name, fd.data, got, err)
				}
			}
		})
	}
}

func TestStorage(t *testing.T) {

	for _, tc := range []struct {
		name    string
		records []*rpb.RecordSet
		want    []*rpb.RecordSet
	}{
		{
			name:    "Empty record set",
			records: []*rpb.RecordSet{{Id: "123"}},
			want:    []*rpb.RecordSet{{Id: "123"}},
		},
		{
			name:    "Record set with request",
			records: []*rpb.RecordSet{newRecordSet(1, "A")},
			want:    []*rpb.RecordSet{newRecordSet(1, "A")},
		},
		{
			name:    "Multiple record sets with requests",
			records: []*rpb.RecordSet{newRecordSet(1, "A"), newRecordSet(2, "B")},
			want:    []*rpb.RecordSet{newRecordSet(1, "A"), newRecordSet(2, "B")},
		},
		{
			name:    "Record set with records",
			records: []*rpb.RecordSet{newRecordSetFull(4, "A", 10)},
			want:    []*rpb.RecordSet{newRecordSetFull(4, "A", 10)},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			s := NewLocalStorage(t.TempDir())
			for _, rec := range tc.records {
				if saveError := s.SaveRecordSet(rec); saveError != nil {
					t.Errorf("Error while saving a request: %s", saveError)
				}
			}
			query := &rpb.Query{Sorting: &rpb.Sorting{Field: rpb.Sorting_FETCH_TIME}}
			fromStorage, readError := s.ListRecordSets(query).Get()
			if readError != nil {
				t.Errorf("Error while reading a request: %s", readError)
				return
			}

			if len(tc.want) != len(fromStorage) {
				t.Errorf("Expected result to be %s, but got %s", tc.want, fromStorage)
				return
			}
			for i, proto := range tc.want {
				if proto.String() != fromStorage[i].String() {
					t.Errorf("Expected result[%d] to be %s, but got %s", i, proto, fromStorage[i])
				}
			}
		})
	}
}
