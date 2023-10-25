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
	"chronicler/webdriver"

	rpb "chronicler/records/proto"
	"github.com/lanseg/golang-commons/collections"
	cm "github.com/lanseg/golang-commons/common"
)

var (
	twitterRe = regexp.MustCompile("(twitter|x.com).*/(?P<twitter_id>[0-9]+)[/]?")
	logger    = cm.NewLogger("main")
)

const (
	webdriverPort  = 2828
	firefoxProfile = "/tmp/tmp.QTFqrzeJX4/"
)

func initWebdriver(scenarios string) webdriver.WebDriver {
	ff := webdriver.StartFirefox(webdriverPort, firefoxProfile)
	driver := ff.Driver
	driver.NewSession()
	sc, err := webdriver.LoadScenarios(scenarios)
	if err != nil {
		logger.Warningf("Cannot load webdriver scenarios from %s: %s", scenarios, err)
	} else {
		logger.Infof("Loaded scenarios from %s", scenarios)
		driver.SetScenarios(sc)
	}
	return driver
}

func findTwitterSource(link string) *rpb.Source {
	matches := collections.NewMap(twitterRe.SubexpNames(), twitterRe.FindStringSubmatch(link))
	if match, ok := matches["twitter_id"]; ok && match != "" {
		return &rpb.Source{
			ChannelId: matches["twitter_id"],
			Type:      rpb.SourceType_TWITTER,
		}
	}
	return nil
}

func linkToTarget(link string) *rpb.Source {
	return &rpb.Source{
		Url:  link,
		Type: rpb.SourceType_WEB,
	}
}

func extractRequests(adapters []adapter.Adapter, rs *rpb.RecordSet) []*rpb.Request {
	result := []*rpb.Request{}
	if len(rs.Records) == 1 && rs.Records[0].Source.Type == rpb.SourceType_WEB {
		return result
	}
	for _, link := range rs.Records[0].Links {
		for _, a := range adapters {
			if target := a.MatchLink(link); target != nil {
				result = append(result, &rpb.Request{
					Id:     rs.Id,
					Target: target,
				})
				break
			}
		}
	}
	return result
}

func main() {
	flag.Parse()

	cfg := GetConfig()
	storage := storage.NewStorage(*cfg.StorageRoot, initWebdriver(*cfg.ScenarioLibrary))

	tgBot := telegram.NewBot(*cfg.TelegramBotKey)
	twClient := twitter.NewClient(*cfg.TwitterApiKey)

	adapters := map[rpb.SourceType]adapter.Adapter{
		rpb.SourceType_TELEGRAM: adapter.NewTelegramAdapter(tgBot),
		rpb.SourceType_TWITTER:  adapter.NewTwitterAdapter(twClient),
		rpb.SourceType_WEB:      adapter.NewWebAdapter(nil),
	}
	linkMatchers := []adapter.Adapter{
		adapters[rpb.SourceType_TWITTER],
		adapters[rpb.SourceType_WEB],
	}

	requests := make(chan *rpb.Request)
	response := make(chan *rpb.Response)
	messages := make(chan *rpb.Message)

	go (func() {
		logger := cm.NewLogger("Chronicler")
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
		logger := cm.NewLogger("Storage")
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
				for _, req := range extractRequests(linkMatchers, records) {
					requests <- req
				}
			}
		}
	})()

	go (func() {
		logger := cm.NewLogger("Messages")
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
		logger := cm.NewLogger("Telegram AutoRequest")
		logger.Infof("Starting telegram periodic fetcher")
		for {
			responses := adapters[rpb.SourceType_TELEGRAM].GetResponse(&rpb.Request{
				Id: cm.UUID4(),
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
