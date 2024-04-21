package main

import (
	"fmt"
	"net/http"
	"net/http/cookiejar"
	"os"
	"time"

	cm "github.com/lanseg/golang-commons/common"
	conc "github.com/lanseg/golang-commons/concurrent"
	tgbot "github.com/lanseg/tgbot"

	"chronicler/adapter"
	pkb_adapter "chronicler/adapter/pikabu"
	tlg_adapter "chronicler/adapter/telegram"
	twi_adapter "chronicler/adapter/twitter"
	web_adapter "chronicler/adapter/web"
	rpb "chronicler/records/proto"
	"chronicler/resolver"
	"chronicler/status"
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
	StatusServerPort  *int    `json:"statusServerPort"`
	StorageServerPort *int    `json:"storageServerPort"`
}

func initHttpClient() *http.Client {
	jar, _ := cookiejar.New(nil)
	return &http.Client{
		Jar: jar,
	}
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

func main() {
	cfg := cm.OrExit(cm.GetConfig[Config](os.Args[1:], "config"))

	logger.Infof("Config.StorageRoot: %s", *cfg.StorageRoot)
	logger.Infof("Config.StorageServerPort: %d", *cfg.StorageServerPort)
	logger.Infof("Config.ScenarioLibrary: %s", *cfg.ScenarioLibrary)
	logger.Infof("Config.TwitterApiKey: %d", len(*cfg.TwitterApiKey))
	logger.Infof("Config.TelegramBotKey: %d", len(*cfg.TelegramBotKey))

	stats := cm.OrExit(status.NewStatusClient(fmt.Sprintf("localhost:%d", *cfg.StatusServerPort)))
	stats.Start()

	storage := cm.OrExit(ep.NewRemoteStorage(fmt.Sprintf("localhost:%d", *cfg.StorageServerPort)))
	webDriver := webdriver.NewBrowser(*cfg.ScenarioLibrary)
	resolver := resolver.NewResolver(webDriver, storage, stats)

	ch := NewLocalChronicler(resolver, storage, stats)
	ch.AddAdapter(rpb.SourceType_TELEGRAM, tlg_adapter.NewTelegramAdapter(tgbot.NewBot(*cfg.TelegramBotKey)))
	ch.AddAdapter(rpb.SourceType_TWITTER, twi_adapter.NewTwitterAdapter(twi_adapter.NewClient(*cfg.TwitterApiKey)))
	ch.AddAdapter(rpb.SourceType_PIKABU, pkb_adapter.NewPikabuAdapter(webDriver))
	ch.AddAdapter(rpb.SourceType_WEB, web_adapter.NewWebAdapter(nil, webDriver))

	ScheduleRepeatedSource(pkb_adapter.NewFreshProvider(initHttpClient()), rpb.WebEngine_HTTP_PLAIN, ch, 2*time.Minute)
	ScheduleRepeatedSource(pkb_adapter.NewHotProvider(initHttpClient()), rpb.WebEngine_WEBDRIVER, ch, 10*time.Minute)
	ScheduleRepeatedSource(pkb_adapter.NewDisputedProvider(initHttpClient()), rpb.WebEngine_WEBDRIVER, ch, 15*time.Minute)

	conc.RunPeriodically(func() {
		ch.SubmitRequest(&rpb.Request{Target: &rpb.Source{Type: rpb.SourceType_TELEGRAM}})
	}, nil, time.Minute)

	ch.Start()
}
