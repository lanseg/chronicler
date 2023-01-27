package main

import (
	"flag"
	"regexp"

	"chronicler"
	"chronicler/storage"
	"chronicler/twitter"
	"chronicler/util"

	rpb "chronicler/proto/records"
)

const (
	twitterApiFlag  = "twitter_api_key"
	storageRootFlag = "storage_root"
)

var (
	twitterApiKey = flag.String(twitterApiFlag, "", "A key for the twitter api.")
	storageRoot   = flag.String(storageRootFlag, "chronicler_storage", "A local folder to save downloads.")
	log           = util.NewLogger("main")
)

type Config struct {
	twitterApiKey *string
	storageRoot   *string
}

func getConfig(configFile string) Config {
	result := Config{}
	result.twitterApiKey = util.Ifnull(twitterApiKey, result.twitterApiKey)
	result.storageRoot = util.Ifnull(storageRoot, result.storageRoot)
	return result
}

func parseRequest(s string) rpb.Request {
	re := regexp.MustCompile("twitter.*/(?P<twitter_id>[0-9]+)[/]?")
	matches := util.NewMap(re.SubexpNames(), re.FindStringSubmatch(s))
	key := s
	if match, ok := matches["twitter_id"]; ok && match != "" {
		key = matches["twitter_id"]
	}
	return rpb.Request{
		RawRequest: s,
		Source: &rpb.Source{
			ChannelId: key,
			Type:      rpb.SourceType_TWITTER,
		},
	}
}

func main() {
	flag.Parse()

	config := getConfig("config")
	chroniclers := map[rpb.SourceType]chronicler.Chronicler{
		rpb.SourceType_TWITTER: chronicler.NewTwitter("twitter", twitter.NewClient(*config.twitterApiKey)),
	}
	stg := storage.NewStorage(*config.storageRoot)

	for _, arg := range flag.Args() {
		request := parseRequest(arg)
		chr, ok := chroniclers[request.Source.Type]
		if !ok {
			log.Warningf("No loader found for request %s", request)
			continue
		}

		conv, err := chr.GetRecords(&request)
		if err != nil {
			log.Errorf("Failed to get conversation for id %s: %s", request, err)
		}
		if err := stg.SaveRecords(request.Source.ChannelId, conv); err != nil {
			log.Warningf("Error while saving a record: %s", err)
		}

	}
}
