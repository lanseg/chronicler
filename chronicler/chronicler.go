package chronicler

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

type Chronicler interface {
	GetRecordSet() *rpb.RecordSet

	SubmitRequest(*rpb.Request)
	SendResponse(*rpb.Response)
}

type ChroniclerImpl struct {
	Chronicler

	logger       *util.Logger
	recordSource RecordSource
	responseSink ResponseSink
	storage      storage.Storage

	requests chan *rpb.Request
	response chan *rpb.Response
	records  chan *rpb.RecordSet
}

func (c *ChroniclerImpl) GetRecordSet() *rpb.RecordSet {
	if c.records == nil {
		c.logger.Warningf("Record source is not configured for this chronicler")
		return nil
	}
	return <-c.records
}

func (c *ChroniclerImpl) SubmitRequest(req *rpb.Request) {
	c.requests <- req
}

func (c *ChroniclerImpl) SendResponse(resp *rpb.Response) {
	if c.response == nil {
		c.logger.Warningf("Response sink is not configured for this chronicler")
		return
	}
	c.response <- resp
}

func (c *ChroniclerImpl) requestLoop() {
	c.logger.Infof("Starting request loop")
	for {
		request := <-c.requests
		c.logger.Infof("New request %s", request)
		records := c.recordSource.GetRequestedRecords(request)
		for _, recordSet := range records {
			c.records <- recordSet
		}
	}
}

func (c *ChroniclerImpl) responseLoop() {
	c.logger.Infof("Starting response loop")
	for {
		response := <-c.response
		c.logger.Infof("New response to %s", response.Source)
		c.responseSink.SendResponse(response)
	}
}

func NewChronicler(recordSrc RecordSource, respSink ResponseSink, loop bool) Chronicler {
	c := &ChroniclerImpl{
		logger:       util.NewLogger("Chronicler"),
		recordSource: recordSrc,
		responseSink: respSink,
		requests:     make(chan *rpb.Request),
	}
	c.logger.Infof("Created new chronicler, loop is %b", loop)
	if recordSrc != nil {
		c.records = make(chan *rpb.RecordSet)
		go c.requestLoop()
	}
	if respSink != nil {
		c.response = make(chan *rpb.Response)
		go c.responseLoop()
	}
	if loop {
		go func() {
			for {
				c.SubmitRequest(nil)
			}
		}()
	}
	return c
}
