package main

import (
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
	StatusServer       *string `json:"statusServer"`
	StorageServer      *string `json:"storageServer"`
	WebdriverServer    *string `json:"webdriverServer"`
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
	logger.Infof("Config.StatusServer: %s", *cfg.StatusServer)
	logger.Infof("Config.StorageRoot: %s", *cfg.StorageRoot)
	logger.Infof("Config.StorageServer: %s", *cfg.StorageServer)
	logger.Infof("Config.TwitterApiKey: %d", len(*cfg.TwitterApiKey))
	logger.Infof("Config.TelegramApiUrl: %s", *cfg.TelegramApiUrl)
	logger.Infof("Config.TelegramBotKey: %d", len(*cfg.TelegramBotKey))
	logger.Infof("Config.TelegramFilePrefix: %s", *cfg.TelegramFilePrefix)

	bot := cm.OrExit(tgbot.NewCustomBot(*cfg.TelegramApiUrl, *cfg.TelegramBotKey))
	storage := cm.OrExit(ep.NewRemoteStorage(*cfg.StorageServer))
	stats := cm.OrExit(status.NewStatusClient(*cfg.StatusServer))
	stats.Start()

	webDriver := webdriver.NewBrowser(*cfg.WebdriverServer, *cfg.ScenarioLibrary)
	resolver := resolver.NewResolver(webDriver, storage, stats)

	ch := chronicler.NewLocalChronicler(resolver, storage, stats)
	ch.AddAdapter(rpb.SourceType_TELEGRAM, tlg_adapter.NewTelegramAdapter(bot, cm.OrExit(url.Parse(*cfg.TelegramFilePrefix))))
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
