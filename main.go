package main

import (
	"crypto/md5"
	"flag"
	"fmt"
	"net/url"
	"regexp"

	"chronicler"
	rpb "chronicler/proto/records"
	"chronicler/storage"
	"chronicler/twitter"
	"chronicler/util"

	"github.com/lanseg/golang-commons/collections"
)

var (
	log = util.NewLogger("main")
)

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
	matches := collections.NewMap(re.SubexpNames(), re.FindStringSubmatch(s))
	if match, ok := matches["twitter_id"]; ok && match != "" {
		source.ChannelId = matches["twitter_id"]
		source.Type = rpb.SourceType_TWITTER
	}
	return rpb.Request{Source: source}
}

func main() {
	flag.Parse()
	cfg := chronicler.GetConfig()
	chroniclers := map[rpb.SourceType]chronicler.Chronicler{
		rpb.SourceType_TWITTER: chronicler.NewTwitter("twitter",
			twitter.NewClient(*cfg.TwitterApiKey)),
		rpb.SourceType_WEB: chronicler.NewWeb("web", nil),
	}
	stg := storage.NewStorage(*cfg.StorageRoot)

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
