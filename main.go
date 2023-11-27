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
	"github.com/lanseg/golang-commons/optional"
)

var (
	twitterRe           = regexp.MustCompile("(twitter|x.com).*/(?P<twitter_id>[0-9]+)[/]?")
	logger              = cm.NewLogger("main")
	telegramRequestUUID = cm.UUID4()
)

func initWebdriver(scenarios string) *webdriver.ExclusiveWebDriver {
	ww, _ := optional.MapErr(webdriver.Connect(),
		func(wd webdriver.WebDriver) (webdriver.WebDriver, error) {
			wd.NewSession()
			sc, err := webdriver.LoadScenarios(scenarios)
			if err != nil {
				logger.Warningf("Cannot load webdriver scenarios from %s: %s", scenarios, err)
				return nil, err
			}
			logger.Infof("Loaded scenarios from %s", scenarios)
			return webdriver.NewScenarioWebdriver(wd, sc), nil
		}).Get()
	return webdriver.WrapExclusive(ww)
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
	if len(rs.Records) > 0 && (rs.Records[0].Source.Type == rpb.SourceType_WEB || rs.Records[0].Source.Type == rpb.SourceType_PIKABU) {
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
	webDriver := initWebdriver(*cfg.ScenarioLibrary)
	storage := storage.NewStorage(*cfg.StorageRoot, webDriver)

	tgBot := telegram.NewBot(*cfg.TelegramBotKey)
	twClient := twitter.NewClient(*cfg.TwitterApiKey)

	adapters := map[rpb.SourceType]adapter.Adapter{
		rpb.SourceType_TELEGRAM: adapter.NewTelegramAdapter(tgBot),
		rpb.SourceType_TWITTER:  adapter.NewTwitterAdapter(twClient),
		rpb.SourceType_PIKABU:   adapter.NewPikabuAdapter(webDriver),
		rpb.SourceType_WEB:      adapter.NewWebAdapter(nil, webDriver),
	}
	linkMatchers := []adapter.Adapter{
		adapters[rpb.SourceType_TWITTER],
		adapters[rpb.SourceType_PIKABU],
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
			} else {
				logger.Infof("No handler for request: %s", newRequest)
			}
		}
	})()

	go (func() {
		logger := cm.NewLogger("Storage")
		logger.Infof("Starting storage thread")
		for {
			result := <-response
			logger.Infof("Got new response for request %s (%s) of size %d", result.Request, result.Request.Origin, len(result.Result))
			for _, records := range result.Result {
				msg := fmt.Sprintf("Saved as %s", records.Id)
				if err := storage.SaveRecords(records); err != nil {
					msg = fmt.Sprintf("Error while saving %q", records.Id)
					logger.Warningf("Error while saving %q: %s", records.Id, err)
				} else {
					logger.Infof("Saved as %q", records.Id)
				}

				if result.Request != nil && result.Request.Origin != nil {
					messages <- &rpb.Message{
						Target:  result.Request.Origin,
						Content: []byte(msg),
					}
				}
			}
			logger.Infof("Extracting requests from %s (%s) of size %d", result.Request,
				result.Request.Origin, len(result.Result))
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
				Id: telegramRequestUUID,
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
