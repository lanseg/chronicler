package chronicler

import (
	rpb "chronicler/proto/records"
)

type Chronicler interface {
	GetName() string
	GetRecords(request *rpb.Request) (*rpb.RecordSet, error)
}
