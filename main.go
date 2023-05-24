package main

import (
	"crypto/md5"
	"encoding/json"
	"flag"
	"fmt"
	"net/url"
	"os"
	"regexp"

	"chronicler"
	"chronicler/storage"
	"chronicler/twitter"
	"chronicler/util"

	rpb "chronicler/proto/records"
)

var (
	configFile    = flag.String("config", "", "Configuration defaults.")
	twitterApiKey = flag.String("twitter_api_key", "", "A key for the twitter api.")
	storageRoot   = flag.String("storage_root", "chronicler_storage", "A local folder to save downloads.")
	log           = util.NewLogger("main")
)

type Config struct {
	TwitterApiKey *string `json:"twitterApiKey"`
	StorageRoot   *string `json:"storageRoot"`
}

func getConfig(configFile *string) Config {
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
	return config
}

func parseRequest(s string) rpb.Request {
	source := &rpb.Source{}
	if parsedUrl, err := url.ParseRequestURI(s); err == nil {
		source.Url = s
		source.ChannelId = fmt.Sprintf("%s_%x",
			parsedUrl.Host,
			md5.Sum([]byte(parsedUrl.String())))
		source.Type = rpb.SourceType_WEB
	}

	re := regexp.MustCompile("twitter.*/(?P<twitter_id>[0-9]+)[/]?")
	matches := util.NewMap(re.SubexpNames(), re.FindStringSubmatch(s))
	if match, ok := matches["twitter_id"]; ok && match != "" {
		source.ChannelId = matches["twitter_id"]
		source.Type = rpb.SourceType_TWITTER
	}
	return rpb.Request{Source: source}
}

func main() {
	flag.Parse()

	config := getConfig(configFile)
	chroniclers := map[rpb.SourceType]chronicler.Chronicler{
		rpb.SourceType_TWITTER: chronicler.NewTwitter("twitter",
			twitter.NewClient(*config.TwitterApiKey)),
		rpb.SourceType_WEB: chronicler.NewWeb("web", nil),
	}
	stg := storage.NewStorage(*config.StorageRoot)

	for _, arg := range flag.Args() {
		request := parseRequest(arg)
		chr, ok := chroniclers[request.Source.Type]
		if !ok {
			log.Warningf("No loader found for request %s", request)
			continue
		}
		conv, err := chr.GetRecords(&request)
		conv.Request = &request
		if err != nil {
			log.Errorf("Failed to get conversation for id %s: %s", request, err)
		}
		if err := stg.SaveRecords(conv); err != nil {
			log.Warningf("Error while saving a record: %s", err)
		}
	}
}
