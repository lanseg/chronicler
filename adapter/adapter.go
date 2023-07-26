package adapter

import (
	"chronicler/storage"
	"chronicler/util"

	rpb "chronicler/proto/records"
)

type RecordSource interface {
	GetRequestedRecords(*rpb.Request) []*rpb.RecordSet
}

type ResponseSink interface {
	SendResponse(*rpb.Response)
}

type SinkSource interface {
	RecordSource
	ResponseSink
}

type Adapter interface {
	GetRecordSet() *rpb.RecordSet

	SubmitRequest(*rpb.Request)
	SendResponse(*rpb.Response)
}

type AdapterImpl struct {
	Adapter

	logger       *util.Logger
	recordSource RecordSource
	responseSink ResponseSink
	storage      storage.Storage

	requests chan *rpb.Request
	response chan *rpb.Response
	records  chan *rpb.RecordSet
}

func (a *AdapterImpl) GetRecordSet() *rpb.RecordSet {
	if a.records == nil {
		a.logger.Warningf("Record source is not configured for this chronicler")
		return nil
	}
	return <-a.records
}

func (a *AdapterImpl) SubmitRequest(req *rpb.Request) {
	a.requests <- req
}

func (a *AdapterImpl) SendResponse(resp *rpb.Response) {
	if a.response == nil {
		a.logger.Warningf("Response sink is not configured for this chronicler")
		return
	}
	a.response <- resp
}

func (a *AdapterImpl) requestLoop() {
	a.logger.Infof("Starting request loop")
	for {
		request := <-a.requests
		a.logger.Infof("New request %s", request)
		records := a.recordSource.GetRequestedRecords(request)
		for _, recordSet := range records {
			a.records <- recordSet
		}
	}
}

func (a *AdapterImpl) responseLoop() {
	a.logger.Infof("Starting response loop")
	for {
		response := <-a.response
		a.logger.Infof("New response to %s", response.Source)
		a.responseSink.SendResponse(response)
	}
}

func NewAdapter(name string, recordSrc RecordSource, respSink ResponseSink, loop bool) Adapter {
	a := &AdapterImpl{
		logger:       util.NewLogger(name),
		recordSource: recordSrc,
		responseSink: respSink,
		requests:     make(chan *rpb.Request),
	}
	a.logger.Infof("Created new adapter, loop is %b", loop)
	if recordSrc != nil {
		a.records = make(chan *rpb.RecordSet)
		go a.requestLoop()
	}
	if respSink != nil {
		a.response = make(chan *rpb.Response)
		go a.responseLoop()
	}
	if loop {
		go func() {
			for {
				a.SubmitRequest(nil)
			}
		}()
	}
	return a
}
