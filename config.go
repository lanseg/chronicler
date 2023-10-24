package main

import (
	"flag"

	"encoding/json"
	"os"

	cm "github.com/lanseg/golang-commons/common"
)

var (
	configFile      = flag.String("config", "", "Configuration defaults.")
	twitterApiKey   = flag.String("twitter_api_key", "", "A key for the twitter api.")
	telegramBotKey  = flag.String("telegram_bot_key", "", "A key for the telegram bot api.")
	storageRoot     = flag.String("storage_root", "", "A local folder to save downloads.")
	scenarioLibrary = flag.String("scenario_library", "", "A list of scripts that adapt webpages for loading.")
	log             = cm.NewLogger("config")
)

type Config struct {
	TwitterApiKey   *string `json:"twitterApiKey"`
	TelegramBotKey  *string `json:"telegramBotKey"`
	StorageRoot     *string `json:"storageRoot"`
	ScenarioLibrary *string `json:"ScenarioLibrary"`
}

func GetConfig() Config {
	flag.Parse()

	config := Config{}
	if configFile != nil {
		b, err := os.ReadFile(*configFile)
		if err != nil {
			log.Warningf("Error reading file %s: %s", config, err)
		} else if err = json.Unmarshal(b, &config); err != nil {
			log.Warningf("Error unmarshalling file %s: %s", config, err)
		}
	}
	if *twitterApiKey != "" {
		config.TwitterApiKey = twitterApiKey
	}
	if *storageRoot != "" {
		config.StorageRoot = storageRoot
	}
	if *telegramBotKey != "" {
		config.TelegramBotKey = telegramBotKey
	}
	if *scenarioLibrary != "" {
		config.ScenarioLibrary = scenarioLibrary
	}
	return config
}
