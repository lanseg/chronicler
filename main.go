package main

import (
	"fmt"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"time"

	cm "github.com/lanseg/golang-commons/common"
	conc "github.com/lanseg/golang-commons/concurrent"
	tgbot "github.com/lanseg/tgbot"

	pkb_adapter "chronicler/adapter/pikabu"
	tlg_adapter "chronicler/adapter/telegram"
	twi_adapter "chronicler/adapter/twitter"
	web_adapter "chronicler/adapter/web"
	"chronicler/chronicler"
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
	TwitterApiKey      *string `json:"twitterApiKey"`
	TelegramApiUrl     *string `json:"telegramApiUrl"`
	TelegramBotKey     *string `json:"telegramBotKey"`
	TelegramFilePrefix *string `json:"telegramFilePrefix"`
	StorageRoot        *string `json:"storageRoot"`
	ScenarioLibrary    *string `json:"scenarioLibrary"`
	StatusServerPort   *int    `json:"statusServerPort"`
	StorageServerPort  *int    `json:"storageServerPort"`
}

func initHttpClient() *http.Client {
	jar, _ := cookiejar.New(nil)
	return &http.Client{
		Jar: jar,
	}
}

func main() {
	cfg := cm.OrExit(cm.GetConfig[Config](os.Args[1:], "config"))

	logger.Infof("Config.ScenarioLibrary: %s", *cfg.ScenarioLibrary)
	logger.Infof("Config.StatusServerPort: %d", *cfg.StatusServerPort)
	logger.Infof("Config.StorageRoot: %s", *cfg.StorageRoot)
	logger.Infof("Config.StorageServerPort: %d", *cfg.StorageServerPort)
	logger.Infof("Config.TwitterApiKey: %d", len(*cfg.TwitterApiKey))
	logger.Infof("Config.TelegramApiUrl: %s", *cfg.TelegramApiUrl)
	logger.Infof("Config.TelegramBotKey: %d", len(*cfg.TelegramBotKey))
	logger.Infof("Config.TelegramFilePrefix: %s", *cfg.TelegramFilePrefix)

	stats := cm.OrExit(status.NewStatusClient(fmt.Sprintf("localhost:%d", *cfg.StatusServerPort)))
	stats.Start()

	storage := cm.OrExit(ep.NewRemoteStorage(fmt.Sprintf("localhost:%d", *cfg.StorageServerPort)))

	webDriver := webdriver.NewBrowser(*cfg.ScenarioLibrary)
	resolver := resolver.NewResolver(webDriver, storage, stats)

	bot := cm.OrExit(tgbot.NewCustomBot(*cfg.TelegramApiUrl, *cfg.TelegramBotKey))
	tgFilePrefix := cm.OrExit(url.Parse(*cfg.TelegramFilePrefix))

	ch := chronicler.NewLocalChronicler(resolver, storage, stats)
	ch.AddAdapter(rpb.SourceType_TELEGRAM, tlg_adapter.NewTelegramAdapter(bot, tgFilePrefix))
	ch.AddAdapter(rpb.SourceType_TWITTER, twi_adapter.NewTwitterAdapter(twi_adapter.NewClient(*cfg.TwitterApiKey)))
	ch.AddAdapter(rpb.SourceType_PIKABU, pkb_adapter.NewPikabuAdapter(webDriver))
	ch.AddAdapter(rpb.SourceType_WEB, web_adapter.NewWebAdapter(nil, webDriver))

	chronicler.ScheduleRepeatedSource(stats, "pikabu.fresh", pkb_adapter.NewFreshProvider(initHttpClient()), rpb.WebEngine_HTTP_PLAIN, ch, 2*time.Minute)
	chronicler.ScheduleRepeatedSource(stats, "pikabu.hot", pkb_adapter.NewHotProvider(initHttpClient()), rpb.WebEngine_WEBDRIVER, ch, 10*time.Minute)
	chronicler.ScheduleRepeatedSource(stats, "pikabu.disputed", pkb_adapter.NewDisputedProvider(initHttpClient()), rpb.WebEngine_WEBDRIVER, ch, 15*time.Minute)

	conc.RunPeriodically(func() {
		stats.PutDateTime("telegram.last_bot_check", time.Now())
		ch.SubmitRequest(&rpb.Request{Target: &rpb.Source{Type: rpb.SourceType_TELEGRAM}})
	}, nil, time.Minute)

	ch.Start()
}
