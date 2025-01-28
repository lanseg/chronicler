package adapter

import (
	opb "chronicler/proto"
)

type Adapter interface {
	Match(link *opb.Link) bool
	Get(link *opb.Link) ([]*opb.Object, error)
}
