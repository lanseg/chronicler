package main

import (
	"fmt"
	"net/http"
	"net/http/cookiejar"
	"os"
	"sync"
	"time"

	cm "github.com/lanseg/golang-commons/common"
	conc "github.com/lanseg/golang-commons/concurrent"
	tgbot "github.com/lanseg/tgbot"

	"chronicler/adapter"
	pkb_adapter "chronicler/adapter/pikabu"
	tlg_adapter "chronicler/adapter/telegram"
	twi_adapter "chronicler/adapter/twitter"
	web_adapter "chronicler/adapter/web"
	"chronicler/downloader"
	rpb "chronicler/records/proto"
	ep "chronicler/storage/endpoint"
	"chronicler/webdriver"
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

func extractRequests(finders []adapter.SourceFinder, rs *rpb.RecordSet) []*rpb.Request {
	result := []*rpb.Request{}
	if len(rs.Records) > 0 && (rs.Records[0].Source.Type == rpb.SourceType_WEB || rs.Records[0].Source.Type == rpb.SourceType_PIKABU) {
		return result
	}
	for _, a := range finders {
		found := false
		for _, target := range a.FindSources(rs.Records[0]) {
			result = append(result, &rpb.Request{
				Id:     rs.Id,
				Target: target,
			})
			found = true
		}
		if found {
			break
		}
	}
	return result
}

func ScheduleRepeatedSource(provider adapter.SourceProvider, engine rpb.WebEngine, ch Chronicler, duration time.Duration) {
	conc.RunPeriodically(func() {
		for _, src := range provider.GetSources() {
			ch.SubmitRequest(&rpb.Request{
				Id: cm.UUID4(),
				Config: &rpb.RequestConfig{
					Engine: engine,
				},
				Target: src,
			})
		}
	}, nil, duration)
}

func sendAll[T any](items []T, ch chan T) {
	for _, item := range items {
		ch <- item
	}
}

func main() {
	cfg := cm.OrExit(cm.GetConfig[Config](os.Args[1:], "config"))

	logger.Infof("Config.StorageRoot: %s", *cfg.StorageRoot)
	logger.Infof("Config.StorageServerPort: %d", *cfg.StorageServerPort)
	logger.Infof("Config.ScenarioLibry: %s", *cfg.ScenarioLibrary)
	logger.Infof("Config.TwitterApiKey: %d", len(*cfg.TwitterApiKey))
	logger.Infof("Config.TelegramBotKey: %d", len(*cfg.TelegramBotKey))

	storage := cm.OrExit(ep.NewRemoteStorage(fmt.Sprintf("localhost:%d", *cfg.StorageServerPort)))

	downloader := downloader.NewDownloader(initHttpClient(), storage)
	webDriver := webdriver.NewBrowser(*cfg.ScenarioLibrary)
	resolver := NewResolver(webDriver, downloader, storage)
	tgBot := tgbot.NewBot(*cfg.TelegramBotKey)
	twClient := twi_adapter.NewClient(*cfg.TwitterApiKey)

	ch := NewLocalChronicler(resolver, storage)
	ch.AddAdapter(rpb.SourceType_TELEGRAM, tlg_adapter.NewTelegramAdapter(tgBot))
	ch.AddAdapter(rpb.SourceType_TWITTER, twi_adapter.NewTwitterAdapter(twClient))
	ch.AddAdapter(rpb.SourceType_PIKABU, pkb_adapter.NewPikabuAdapter(webDriver))
	ch.AddAdapter(rpb.SourceType_WEB, web_adapter.NewWebAdapter(nil, webDriver))

	ScheduleRepeatedSource(pkb_adapter.NewFreshProvider(initHttpClient()), rpb.WebEngine_HTTP_PLAIN, ch, 5*time.Minute)
	ScheduleRepeatedSource(pkb_adapter.NewHotProvider(initHttpClient()), rpb.WebEngine_WEBDRIVER, ch, 15*time.Minute)
	ScheduleRepeatedSource(pkb_adapter.NewDisputedProvider(initHttpClient()), rpb.WebEngine_WEBDRIVER, ch, 30*time.Minute)

	conc.RunPeriodically(func() {
		ch.SubmitRequest(&rpb.Request{Target: &rpb.Source{Type: rpb.SourceType_TELEGRAM}})
	}, nil, time.Minute)

	ch.Start()

	wg := sync.WaitGroup{}
	wg.Add(1)
	wg.Wait()
}
