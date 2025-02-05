package adapter

import (
	"net/http"

	opb "chronicler/proto"
)

type HttpClient interface {
	Do(request *http.Request) (*http.Response, error)
}

type Adapter interface {
	Match(link *opb.Link) bool
	Get(link *opb.Link) ([]*opb.Object, error)
}
