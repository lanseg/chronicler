package adapter

import (
	rpb "chronicler/records/proto"
)

type Adapter interface {
	GetResponse(*rpb.Request) []*rpb.Response
	SendMessage(*rpb.Message)
}
