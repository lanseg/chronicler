package main

import (
	"fmt"
	"net/http"
	"net/http/cookiejar"
	"os"
	"sync"

	"chronicler/adapter"
	"chronicler/downloader"
	"chronicler/storage"
	"chronicler/twitter"
	"chronicler/webdriver"

	rpb "chronicler/records/proto"

	cm "github.com/lanseg/golang-commons/common"
	tgbot "github.com/lanseg/tgbot"
)

var (
	logger              = cm.NewLogger("main")
	telegramRequestUUID = cm.UUID4()
)

type Config struct {
	TwitterApiKey     *string `json:"twitterApiKey"`
	TelegramBotKey    *string `json:"telegramBotKey"`
	StorageRoot       *string `json:"storageRoot"`
	ScenarioLibrary   *string `json:"scenarioLibrary"`
	StorageServerPort *int    `json:"storageServerPort"`
}

func initHttpClient() *http.Client {
	jar, _ := cookiejar.New(nil)
	return &http.Client{
		Jar: jar,
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
	cfg, err := cm.GetConfig[Config](os.Args[1:], "config")
	if err != nil {
		logger.Errorf("Could not load config: %s", err)
		os.Exit(-1)
	}
	logger.Infof("Config.StorageRoot: %s", *cfg.StorageRoot)
	logger.Infof("Config.StorageServerPort: %d", *cfg.StorageServerPort)
	logger.Infof("Config.ScenarioLibry: %s", *cfg.ScenarioLibrary)
	logger.Infof("Config.TwitterApiKey: %d", len(*cfg.TwitterApiKey))
	logger.Infof("Config.TelegramBotKey: %d", len(*cfg.TelegramBotKey))

	downloader := downloader.NewDownloader(initHttpClient())
	webDriver := webdriver.NewBrowser(*cfg.ScenarioLibrary)
	storage := storage.NewStorage(*cfg.StorageRoot, webDriver, downloader)

	tgBot := tgbot.NewBot(*cfg.TelegramBotKey)
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

	requests := make(chan *rpb.Request, 10)
	response := make(chan *rpb.Response, 10)
	messages := make(chan *rpb.Message, 10)

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
				if err := storage.SaveRecordSet(records); err != nil {
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
					req.Origin = result.Request.Origin
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
