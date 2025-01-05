package storage

import (
	"io"

	opt "github.com/lanseg/golang-commons/optional"

	rpb "chronicler/records/proto"
)

const (
	recordsetFileName = "record.json"
)

type Storage interface {
	SaveRecordSet(*rpb.RecordSet) error
	ListRecordSets(*rpb.ListRecordsRequest) opt.Optional[[]*rpb.RecordSet]
	GetRecordSet(id string) opt.Optional[*rpb.RecordSet]
	DeleteRecordSet(id string) error

	GetFile(id string, filename string) opt.Optional[io.ReadCloser]
	PutFile(id string, filename string, src io.Reader) error
}

type NoOpStorage struct {
	Storage
}

func (ns *NoOpStorage) SaveRecordSet(r *rpb.RecordSet) error {
	return nil
}

func (ns *NoOpStorage) ListRecordSets(*rpb.ListRecordsRequest) opt.Optional[[]*rpb.RecordSet] {
	return opt.Nothing[[]*rpb.RecordSet]{}
}

func (ns *NoOpStorage) GetRecordSet(id string) opt.Optional[*rpb.RecordSet] {
	return opt.Nothing[*rpb.RecordSet]{}
}

func (ns *NoOpStorage) DeleteRecordSet(id string) error {
	return nil
}

func (ns *NoOpStorage) GetFile(id string, filename string) opt.Optional[io.ReadCloser] {
	return opt.Nothing[io.ReadCloser]{}
}

func (ns *NoOpStorage) PutFile(id string, filename string, src io.Reader) error {
	return nil
}
