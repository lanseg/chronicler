package main

import (
	"flag"
	"fmt"
	"regexp"
	"sync"

	"chronicler/adapter"
	"chronicler/storage"
	"chronicler/telegram"
	"chronicler/twitter"
	"chronicler/util"

	"github.com/lanseg/golang-commons/collections"

	rpb "chronicler/records/proto"
)

var (
	twitterRe = regexp.MustCompile("twitter.*/(?P<twitter_id>[0-9]+)[/]?")
)

func extractRequests(log *util.Logger, rs *rpb.RecordSet) []*rpb.Request {
	result := []*rpb.Request{}
	if len(rs.Records) == 1 && rs.Records[0].Source.Type == rpb.SourceType_WEB {
		return []*rpb.Request{}
	}
	for _, link := range rs.Records[0].Links {
		matches := collections.NewMap(twitterRe.SubexpNames(), twitterRe.FindStringSubmatch(link))
		newRequest := &rpb.Request{
			Id: rs.Id,
		}

		if match, ok := matches["twitter_id"]; ok && match != "" {
			newRequest.Target = &rpb.Source{
				ChannelId: matches["twitter_id"],
				Type:      rpb.SourceType_TWITTER,
			}
		} else {
			newRequest.Target = &rpb.Source{
				Url:  link,
				Type: rpb.SourceType_WEB,
			}
		}
		result = append(result, newRequest)
	}
	return result
}

func main() {
	flag.Parse()

	cfg := GetConfig()
	storage := storage.NewStorage(*cfg.StorageRoot, nil)

	tgBot := telegram.NewBot(*cfg.TelegramBotKey)
	twClient := twitter.NewClient(*cfg.TwitterApiKey)

	adapters := map[rpb.SourceType]adapter.Adapter{
		rpb.SourceType_TELEGRAM: adapter.NewTelegramAdapter(tgBot),
		rpb.SourceType_TWITTER:  adapter.NewTwitterAdapter(twClient),
		rpb.SourceType_WEB:      adapter.NewWebAdapter(nil),
	}

	requests := make(chan *rpb.Request)
	response := make(chan *rpb.Response)
	messages := make(chan *rpb.Message)

	go (func() {
		logger := util.NewLogger("Chronicler")
		logger.Infof("Starting chronicler thread")
		for {
			newRequest := <-requests
			logger.Infof("Got new request: %s", newRequest)
			if a, ok := adapters[newRequest.Target.Type]; ok {
				for _, resp := range a.GetResponse(newRequest) {
					response <- resp
				}
			}
			logger.Infof("No handler for request: %s", newRequest)
		}
	})()

	go (func() {
		logger := util.NewLogger("Storage")
		logger.Infof("Starting storage thread")
		for {
			result := <-response
			logger.Infof("Got new response for request %s", result.Request)
			for _, records := range result.Result {
				msg := fmt.Sprintf("Saved as %s", records.Id)
				if err := storage.SaveRecords(records); err != nil {
					msg = fmt.Sprintf("Error while saving %s", records.Id)
				}
				if result.Request != nil && result.Request.Origin != nil {
					messages <- &rpb.Message{
						Target:  result.Request.Origin,
						Content: []byte(msg),
					}
				}
			}
			for _, records := range result.Result {
				for _, req := range extractRequests(logger, records) {
					requests <- req
				}
			}
		}
	})()

	go (func() {
		logger := util.NewLogger("Messages")
		logger.Infof("Starting message thread")
		for {
			newMessage := <-messages
			logger.Infof("Got new message to send: %s", newMessage)
			if a, ok := adapters[newMessage.Target.Type]; ok {
				a.SendMessage(newMessage)
			}
			logger.Infof("No adapter for message %s", newMessage)
		}
	})()

	go (func() {
		logger := util.NewLogger("Telegram AutoRequest")
		logger.Infof("Starting telegram periodic fetcher")
		for {
			responses := adapters[rpb.SourceType_TELEGRAM].GetResponse(&rpb.Request{
				Id: util.UUID4(),
			})
			if len(responses) == 0 {
				continue
			}
			for _, resp := range responses {
				response <- resp
			}
		}
	})()
	wg := sync.WaitGroup{}
	wg.Add(1)
	wg.Wait()
}
