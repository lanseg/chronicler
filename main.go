package main

import (
	"flag"
	"net/url"
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
	source := &rpb.Source{}
	if parsedUrl, err := url.ParseRequestURI(s); err == nil {
		source.Url = s
		source.ChannelId = parsedUrl.Host
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

	config := getConfig("config")
	chroniclers := map[rpb.SourceType]chronicler.Chronicler{
		rpb.SourceType_TWITTER: chronicler.NewTwitter("twitter", twitter.NewClient(*config.twitterApiKey)),
		rpb.SourceType_WEB:     chronicler.NewWeb("web", nil),
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
		conv.Request = &request
		if err != nil {
			log.Errorf("Failed to get conversation for id %s: %s", request, err)
		}
		if err := stg.SaveRecords(conv); err != nil {
			log.Warningf("Error while saving a record: %s", err)
		}
	}
}
