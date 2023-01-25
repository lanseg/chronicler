package chronist

import (
	rpb "chronist/proto/records"
)

type Chronist interface {
	GetName()
	GetRecords(request *rpb.Request) []*rpb.Record
}
