package storage

import (
	"fmt"
	"testing"

	rpb "chronicler/proto/records"
)

func newRecordSet(name string) *rpb.RecordSet {
	rs := &rpb.RecordSet{
		Request: &rpb.Request{
			Source: &rpb.Source{
				SenderId:  "SenderId" + name,
				ChannelId: "ChannelId" + name,
				MessageId: "MessageId" + name,
				Url:       "http://url.domain/" + name,
			},
		},
	}
	rs.Id = getRecordSetId(rs)
	return rs
}

func newRecordSetFull(name string, nrecords int) *rpb.RecordSet {
	set := newRecordSet(name)
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
			records: []*rpb.RecordSet{newRecordSet("A")},
			want:    []*rpb.RecordSet{newRecordSet("A")},
		},
		{
			name:    "Multiple record sets with requests",
			records: []*rpb.RecordSet{newRecordSet("A"), newRecordSet("B")},
			want:    []*rpb.RecordSet{newRecordSet("A"), newRecordSet("B")},
		},
		{
			name:    "Record set with records",
			records: []*rpb.RecordSet{newRecordSetFull("A", 10)},
			want:    []*rpb.RecordSet{newRecordSetFull("A", 10)},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			s := NewStorage(t.TempDir())
			for _, rec := range tc.records {
				if saveError := s.SaveRecords(rec); saveError != nil {
					t.Errorf("Error while saving a request: %s", saveError)
				}
			}
			fromStorage, readError := s.ListRecords()
			if readError != nil {
				t.Errorf("Error while reading a request: %s", readError)
			}

			if len(tc.want) != len(fromStorage) {
				t.Errorf("Expected result to be %s, but got %s", tc.want, fromStorage)
			}
			for i, proto := range tc.want {
				if proto.String() != fromStorage[i].String() {
					t.Errorf("Expected result[%d] to be %s, but got %s", i, proto, fromStorage[i])
				}
			}
		})
	}
}
