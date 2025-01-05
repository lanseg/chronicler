package adapter

import (
	rpb "chronicler/records/proto"
)

type Adapter interface {
	SourceFinder
	ResponseProvider
	MessageSender
}

type SourceFinder interface {
	FindSources(*rpb.Record) []*rpb.Source
}

type SourceProvider interface {
	GetSources() []*rpb.Source
}

type ResponseProvider interface {
	GetResponse(*rpb.Request) []*rpb.Response
}

type MessageSender interface {
	SendMessage(*rpb.Message)
}
