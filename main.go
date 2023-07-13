package main

import (
	"flag"
	"fmt"
	"regexp"
	"sync"

	"chronicler"
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
	if len(rs.Records) != 1 {
		log.Debugf(
			"We expect exactly 1 record in RecordSet, but got %d when extracting requests",
			len(rs.Records))
		return result
	}
	for _, link := range rs.Records[0].Links {
		matches := collections.NewMap(twitterRe.SubexpNames(), twitterRe.FindStringSubmatch(link))
		if match, ok := matches["twitter_id"]; ok && match != "" {
			result = append(result, &rpb.Request{
				Parent: rs.Request.Source,
				Source: &rpb.Source{
					ChannelId: matches["twitter_id"],
					Type:      rpb.SourceType_TWITTER,
				},
			})
		}
	}
	return result
}

func main() {
	flag.Parse()
	cfg := chronicler.GetConfig()
	stg := storage.NewStorage(*cfg.StorageRoot)
	chroniclers := map[rpb.SourceType]chronicler.Chronicler{
		rpb.SourceType_TELEGRAM: chronicler.NewTelegramChronicler(
			telegram.NewBot(*cfg.TelegramBotKey)),
		rpb.SourceType_TWITTER: chronicler.NewTwitterChronicler(
			twitter.NewClient(*cfg.TwitterApiKey)),
	}

	for srcType, chr := range chroniclers {
		go func(srcType rpb.SourceType, chr chronicler.Chronicler) {
			log := util.NewLogger(fmt.Sprintf("%s Record loader", srcType))
			recordSource := chr.GetRecordSource()
			for {
				recordSet := <-recordSource
				for _, newRequest := range extractRequests(log, recordSet) {
					if targetChr, ok := chroniclers[newRequest.Source.Type]; ok {
						targetChr.SubmitRequest(newRequest)
					}
				}
				src := recordSet.Request.Source
				if err := stg.SaveRecords(recordSet); err != nil {
					log.Warningf("Error while saving a record: %s", err)
					chr.SendResponse(&rpb.Response{Source: src, Content: err.Error()})
				} else {
					chr.SendResponse(&rpb.Response{Source: src, Content: "Saved"})
				}
			}
		}(srcType, chr)
	}

	wg := sync.WaitGroup{}
	wg.Add(1)
	wg.Wait()
}
