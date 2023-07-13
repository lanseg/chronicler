package chronicler

import (
	rpb "chronicler/proto/records"
)

type Chronicler interface {
	GetRecordSource() <-chan *rpb.RecordSet

	SubmitRequest(*rpb.Request)
	SendResponse(*rpb.Response)
}
