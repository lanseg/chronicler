package storage

import (
	"fmt"
	"testing"

	"chronicler/downloader"
	rpb "chronicler/records/proto"
	"chronicler/webdriver"
)

type fakeDownloader struct {
	downloader.Downloader
}

func (fd *fakeDownloader) ScheduleDownload(string, string) error {
	return nil
}

type FakeDriver struct {
	webdriver.WebDriver
}

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
			s := NewStorage(t.TempDir(), &webdriver.ExclusiveWebDriver{}, &fakeDownloader{})
			for _, rec := range tc.records {
				if saveError := s.SaveRecordSet(rec); saveError != nil {
					t.Errorf("Error while saving a request: %s", saveError)
				}
			}
			fromStorage, readError := s.ListRecordSets().Get()
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
