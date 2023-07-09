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

type ResponseSender interface {
	SendResponse() chan<- *rpb.Response
}

type RequestRecordSource interface {
	RequestSource
	RecordSource
}

type Chronicler interface {
	RequestSource
	RecordSource
	ResponseSender
}
