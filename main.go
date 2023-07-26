package main

import (
	"flag"
	"fmt"
	"regexp"
	"sync"

	"chronicler/adapter"
	"chronicler/storage"
	"chronicler/telegram"
	"chronicler/twitter"
	"chronicler/util"

	"github.com/lanseg/golang-commons/collections"

	rpb "chronicler/proto/records"
)

var (
	twitterRe = regexp.MustCompile("twitter.*/(?P<twitter_id>[0-9]+)[/]?")
)

func extractRequests(log *util.Logger, rs *rpb.RecordSet) []*rpb.Request {
	result := []*rpb.Request{}
	if len(rs.Records) != 1 || rs.Request.Source.Type == rpb.SourceType_WEB {
		log.Debugf(
			"Expected 1 non-web record in RecordSet, but got %d when extracting requests",
			len(rs.Records))
		return result
	}
	for _, link := range rs.Records[0].Links {
		matches := collections.NewMap(twitterRe.SubexpNames(), twitterRe.FindStringSubmatch(link))
		if match, ok := matches["twitter_id"]; ok && match != "" {
			result = append(result, &rpb.Request{
				Parent: rs.Request.Parent,
				Source: &rpb.Source{
					ChannelId: matches["twitter_id"],
					Type:      rpb.SourceType_TWITTER,
				},
			})
		} else {
			result = append(result, &rpb.Request{
				Parent: rs.Request.Source,
				Source: &rpb.Source{
					Url:  link,
					Type: rpb.SourceType_WEB,
				},
			})
		}
	}
	return result
}

func main() {
	flag.Parse()

	cfg := GetConfig()
	stg := storage.NewStorage(*cfg.StorageRoot)
	adapters := map[rpb.SourceType]adapter.Adapter{
		rpb.SourceType_TELEGRAM: adapter.NewTelegramAdapter(
			telegram.NewBot(*cfg.TelegramBotKey)),
		rpb.SourceType_TWITTER: adapter.NewTwitterAdapter(
			twitter.NewClient(*cfg.TwitterApiKey)),
		rpb.SourceType_WEB: adapter.NewWebAdapter(nil),
	}

	for srcType, chr := range adapters {
		go func(srcType rpb.SourceType, chr adapter.Adapter) {
			log := util.NewLogger(fmt.Sprintf("%s Record loader", srcType))
			for {
				recordSet := chr.GetRecordSet()
				for _, newRequest := range extractRequests(log, recordSet) {
					if targetChr, ok := adapters[newRequest.Source.Type]; ok {
						targetChr.SubmitRequest(newRequest)
					}
				}

				err := stg.SaveRecords(recordSet)
				responseMessage := "Saved"
				if err != nil {
					responseMessage = "Error"
					log.Warningf("Error while saving a record: %s", err)
				}

				src := recordSet.Request
				if src.Source.Type == rpb.SourceType_TELEGRAM {
					chr.SendResponse(&rpb.Response{Source: src.Source, Content: responseMessage})
				}
			}
		}(srcType, chr)
	}

	wg := sync.WaitGroup{}
	wg.Add(1)
	wg.Wait()
}
