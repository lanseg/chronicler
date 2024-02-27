package adapter

import (
	rpb "chronicler/records/proto"
)

type Adapter interface {
	FindSources(*rpb.Record) []*rpb.Source
	GetResponse(*rpb.Request) []*rpb.Response
	SendMessage(*rpb.Message)
}
