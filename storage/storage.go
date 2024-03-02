package storage

import (
	"io"

	rpb "chronicler/records/proto"

	"github.com/lanseg/golang-commons/optional"
)

const (
	recordsetFileName = "record.json"
)

type Storage interface {
	SaveRecordSet(r *rpb.RecordSet) error
	ListRecordSets() optional.Optional[[]*rpb.RecordSet]
	GetRecordSet(id string) optional.Optional[*rpb.RecordSet]
	DeleteRecordSet(id string) error

	GetFile(id string, filename string) optional.Optional[io.ReadCloser]
	PutFile(id string, filename string, src io.Reader) error
}

type NoOpStorage struct {
	Storage
}

func (ns *NoOpStorage) SaveRecordSet(r *rpb.RecordSet) error {
	return nil
}

func (ns *NoOpStorage) ListRecordSets() optional.Optional[[]*rpb.RecordSet] {
	return optional.Nothing[[]*rpb.RecordSet]{}
}

func (ns *NoOpStorage) GetRecordSet(id string) optional.Optional[*rpb.RecordSet] {
	return optional.Nothing[*rpb.RecordSet]{}
}

func (ns *NoOpStorage) DeleteRecordSet(id string) error {
	return nil
}

func (ns *NoOpStorage) GetFile(id string, filename string) optional.Optional[io.ReadCloser] {
	return optional.Nothing[io.ReadCloser]{}
}

func (ns *NoOpStorage) PutFile(id string, filename string, src io.Reader) error {
	return nil
}
