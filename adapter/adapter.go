package adapter

import (
	opb "chronicler/proto"
)

type Adapter interface {
	Get(link *opb.Link) ([]*opb.Object, error)
}
