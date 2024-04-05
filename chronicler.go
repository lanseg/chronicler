package main

import (
	"fmt"
	"strings"
	"sync"
	"sync/atomic"

	cm "github.com/lanseg/golang-commons/common"

	"chronicler/adapter"
	rpb "chronicler/records/proto"
	"chronicler/storage"
)

type ChroniclerStatus struct {
	waiter   sync.WaitGroup
	jobCount uint32
}

func (cs *ChroniclerStatus) StartJob() {
	cs.waiter.Add(1)
	atomic.AddUint32(&cs.jobCount, uint32(1))
}

func (cs *ChroniclerStatus) StopJob() {
	cs.waiter.Done()
	atomic.AddUint32(&cs.jobCount, ^uint32(0))
}

func (cs *ChroniclerStatus) Wait() {
	cs.waiter.Wait()
}

type Chronicler interface {
	AddAdapter(rpb.SourceType, adapter.Adapter)
	AddResponseProvider(rpb.SourceType, adapter.ResponseProvider)
	AddSourceFinder(rpb.SourceType, adapter.SourceFinder)
	AddMessageSender(rpb.SourceType, adapter.MessageSender)

	SubmitRequest(*rpb.Request)

	Start()
	Stop()
}

type localChronicler struct {
	Chronicler

	logger   *cm.Logger
	resolver Resolver
	storage  storage.Storage
	status   *ChroniclerStatus
	done     chan bool

	requests chan *rpb.Request
	response chan *rpb.Response
	messages chan *rpb.Message

	responseProviders map[rpb.SourceType]adapter.ResponseProvider
	sourceFinders     map[rpb.SourceType]adapter.SourceFinder
	messageSenders    map[rpb.SourceType]adapter.MessageSender
}

func NewLocalChronicler(resolver Resolver, storage storage.Storage) Chronicler {
	return &localChronicler{
		logger:   cm.NewLogger("chronicler"),
		resolver: resolver,
		storage:  storage,
		status:   &ChroniclerStatus{},

		done:     make(chan bool),
		requests: make(chan *rpb.Request),
		response: make(chan *rpb.Response),
		messages: make(chan *rpb.Message),

		responseProviders: map[rpb.SourceType]adapter.ResponseProvider{},
		sourceFinders:     map[rpb.SourceType]adapter.SourceFinder{},
		messageSenders:    map[rpb.SourceType]adapter.MessageSender{},
	}
}

func (ch *localChronicler) AddAdapter(srctype rpb.SourceType, a adapter.Adapter) {
	ch.logger.Infof("Adding adapter of type %q", srctype)
	ch.AddResponseProvider(srctype, a)
	ch.AddSourceFinder(srctype, a)
	ch.AddMessageSender(srctype, a)
}

func (ch *localChronicler) AddResponseProvider(srctype rpb.SourceType, provider adapter.ResponseProvider) {
	ch.logger.Infof("Adding new response provider of %s type", srctype)
	ch.responseProviders[srctype] = provider
}

func (ch *localChronicler) AddSourceFinder(srctype rpb.SourceType, finder adapter.SourceFinder) {
	ch.logger.Infof("Adding new source finder of %s type", srctype)
	ch.sourceFinders[srctype] = finder
}

func (ch *localChronicler) AddMessageSender(srctype rpb.SourceType, sender adapter.MessageSender) {
	ch.logger.Infof("Adding new message sender of %s type", srctype)
	ch.messageSenders[srctype] = sender
}

func (ch *localChronicler) SendMessage(msg *rpb.Message) {
	if a, ok := ch.messageSenders[msg.Target.Type]; ok {
		a.SendMessage(msg)
	} else {
		ch.logger.Infof("No handler for message: %s", msg)
	}
}

func (ch *localChronicler) SubmitRequest(newRequest *rpb.Request) {
	if provider, ok := ch.responseProviders[newRequest.Target.Type]; ok {
		for _, resp := range provider.GetResponse(newRequest) {
			ch.response <- resp
		}
	} else {
		logger.Infof("No handler for request: %s", newRequest)
	}
}

func (ch *localChronicler) HandleResponse(resp *rpb.Response) {
	report := make([]string, len(resp.Result))
	for i, rs := range resp.Result {
		if err := ch.storage.SaveRecordSet(rs); err != nil {
			report[i] = fmt.Sprintf("Error while saving %q: %s", rs.Id, err)
			logger.Warningf(report[i])
		} else {
			report[i] = fmt.Sprintf("Saved as %s", rs.Id)
			logger.Infof(report[i])
		}
	}

	if resp.Request != nil && resp.Request.Origin != nil {
		ch.messages <- &rpb.Message{
			Target:  resp.Request.Origin,
			Content: []byte(strings.Join(report, "\n")),
		}
	}
}

func (ch *localChronicler) ResolveRecordSet(rs *rpb.RecordSet) {
	if err := ch.resolver.Resolve(rs.Id); err != nil {
		logger.Warningf("Error while resolving record contents for %s: %s", rs.Id, err)
	}
}

func (ch *localChronicler) FindSources(resp *rpb.Response) {
	result := []*rpb.Request{}
	for _, rs := range resp.Result {
		if len(rs.Records) > 0 && (rs.Records[0].Source.Type == rpb.SourceType_WEB || rs.Records[0].Source.Type == rpb.SourceType_PIKABU) {
			continue
		}
		for _, record := range rs.Records {
			for _, a := range ch.sourceFinders {
				found := false
				for _, target := range a.FindSources(record) {
					result = append(result, &rpb.Request{
						Id:     rs.Id,
						Target: target,
					})
					found = true
				}
				if found {
					continue
				}
			}
		}
	}
	for _, rq := range result {
		ch.requests <- rq
	}
}

// ---
// ---
func (ch *localChronicler) closeChannels() {
	ch.logger.Infof("Closing channels")
	close(ch.requests)
	close(ch.messages)
	close(ch.response)
	close(ch.done)
}

func (ch *localChronicler) Start() {
	ch.logger.Infof("Starting chronicler")
	ch.status.StartJob()
	go func() {
		done := false
		for !done {
			select {
			case <-ch.done:
				ch.logger.Infof("Shutting down chronicler gracefully")
				done = true
			case msg := <-ch.messages:
				go func() {
					ch.status.StartJob()
					defer ch.status.StopJob()

					ch.SendMessage(msg)
				}()
			case request := <-ch.requests:
				go func() {
					ch.status.StartJob()
					defer ch.status.StopJob()

					ch.SubmitRequest(request)
				}()
			case response := <-ch.response:
				go func() {
					ch.status.StartJob()
					defer ch.status.StopJob()

					ch.HandleResponse(response)
					ch.FindSources(response)
					for _, rs := range response.Result {
						ch.ResolveRecordSet(rs)
					}
				}()
			}
			ch.logger.Infof("Job count: %d", atomic.LoadUint32(&ch.status.jobCount))
		}
		ch.status.StopJob()
		ch.closeChannels()
		ch.logger.Infof("Stopped chronicler")
	}()
	ch.logger.Infof("Started chronicler")
	ch.status.Wait()
}

func (ch *localChronicler) Stop() {
	ch.done <- true
}
