package chronist

import (
	rpb "chronist/proto/records"
)

const (
    chronistName = "Twitter chronicler"
)

type Twitter interface {
    Chronist

	GetRecords(request *rpb.Request) []*rpb.Record
}

func (t *Twitter) GetName() string {
    return chronistName
}

func (t *Twitter) GetRecords(request *rpb.Request) []*rpb.Record {
  return []*rpb.Record{}
} 
