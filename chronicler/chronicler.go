package chronicler

import (
	rpb "chronicler/proto/records"
)

type Fetcher func(*rpb.Request) []*rpb.Response

type Chronicler interface {
	GetRecordSet() *rpb.RecordSet

	SubmitRequest(*rpb.Request)
	SendResponse(*rpb.Response)
}

type ChroniclerImpl struct {
	Chronicler

	requests chan *rpb.Request
	response chan *rpb.Response
	records  chan *rpb.RecordSet
}

func (c *ChroniclerImpl) GetRecordSource() *rpb.RecordSet {
	return <-c.records
}

func (c *ChroniclerImpl) SubmitRequest(req *rpb.Request) {
	c.requests <- req
}

func (c *ChroniclerImpl) SendResponse(resp *rpb.Response) {
	c.response <- resp
}
