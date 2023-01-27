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

	var twt chronicler.Chronicler = chronicler.NewTwitter("twitter", twitter.NewClient(*twitterApiKey))
	stg := storage.NewStorage(*storageRoot)

	for _, arg := range flag.Args() {
		request := parseRequest(arg)
		switch srcType := request.Source.Type; srcType {
		case rpb.SourceType_TWITTER:
			conv, err := twt.GetRecords(&request)
			if err != nil {
				log.Errorf("Failed to get conversation for id %s: %s", request, err)
			}
			if err := stg.SaveRecords(request.Source.ChannelId, conv); err != nil {
				log.Warningf("Error while saving a record: %s", err)
			}
		default:
			log.Warningf("No loader found for request %s", request)
		}
	}
}
