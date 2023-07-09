package chronicler

import (
	rpb "chronicler/proto/records"
)

type RequestSource interface {
	GetRequest() <-chan *rpb.Request
}

type RecordSource interface {
	GetRecords() <-chan *rpb.RecordSet
}

type RequestRecordSource interface {
	RequestSource
	RecordSource
}

type Chronicler interface {
	GetName() string
	GetRecords(request *rpb.Request) (*rpb.RecordSet, error)
}
