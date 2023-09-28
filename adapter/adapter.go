package adapter

import (
	"chronicler/storage"
	"chronicler/util"

	rpb "chronicler/records/proto"
)

type RecordSource interface {
	GetRequestedRecords(*rpb.Request) []*rpb.RecordSet
}

type MessageSink interface {
	SendMessage(*rpb.Message)
}

type SinkSource interface {
	RecordSource
	MessageSink
}

type Adapter interface {
	GetResponse() *rpb.Response
	SubmitRequest(*rpb.Request)
	SendMessage(*rpb.Message)
}

type AdapterImpl struct {
	Adapter

	logger       *util.Logger
	recordSource RecordSource
	messageSink  MessageSink
	storage      storage.Storage

	requests chan *rpb.Request
	messages chan *rpb.Message
	response chan *rpb.Response
}

func (a *AdapterImpl) GetResponse() *rpb.Response {
	if a.response == nil {
		a.logger.Warningf("Record source is not configured for this chronicler")
		return nil
	}
	return <-a.response
}

func (a *AdapterImpl) SubmitRequest(req *rpb.Request) {
	a.requests <- req
}

func (a *AdapterImpl) SendMessage(msg *rpb.Message) {
	if a.messages == nil {
		a.logger.Warningf("Message sink is not configured for this chronicler")
		return
	}
	a.messages <- msg
}

func (a *AdapterImpl) requestResponseLoop() {
	a.logger.Infof("Starting request response loop")
	for {
		request := <-a.requests
		if request == nil {
			a.logger.Warningf("Got empty request, skipping fetch")
			continue
		}

		recordSets := a.recordSource.GetRequestedRecords(request)
		if len(recordSets) == 0 {
			a.logger.Warningf("Empty response for request %v", request)
		}
		a.response <- &rpb.Response{
			Request: request,
			Result:  recordSets,
		}
	}
}

func (a *AdapterImpl) messageLoop() {
	a.logger.Infof("Starting message loop")
	for {
		msg := <-a.messages
		a.logger.Infof("New message to %s", msg.Target)
		a.messageSink.SendMessage(msg)
	}
}

func NewAdapter(name string, recordSrc RecordSource, msgSink MessageSink, loop bool) Adapter {
	a := &AdapterImpl{
		logger:       util.NewLogger(name),
		recordSource: recordSrc,
		messageSink:  msgSink,
		requests:     make(chan *rpb.Request),
	}
	a.logger.Infof("Created new adapter, loop is %v", loop)
	if recordSrc != nil {
		a.response = make(chan *rpb.Response)
		go a.requestResponseLoop()
	}
	if msgSink != nil {
		a.messages = make(chan *rpb.Message)
		go a.messageLoop()
	}
	if loop {
		go func() {
			for {
				a.SubmitRequest(&rpb.Request{
					Id: util.UUID4(),
				})
			}
		}()
	}
	return a
}
