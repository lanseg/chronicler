package adapter

import (
	rpb "chronicler/records/proto"
)

type Adapter interface {
	MatchLink(link string) *rpb.Source
	GetResponse(*rpb.Request) []*rpb.Response
	SendMessage(*rpb.Message)
}
