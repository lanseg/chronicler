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

	rpb "chronicler/records/proto"
)

var (
	twitterRe = regexp.MustCompile("twitter.*/(?P<twitter_id>[0-9]+)[/]?")
)

func extractRequests(log *util.Logger, rs *rpb.RecordSet) []*rpb.Request {
	result := []*rpb.Request{}
	if len(rs.Records) == 1 && rs.Records[0].Source.Type == rpb.SourceType_WEB {
		return []*rpb.Request{}
	}
	for _, link := range rs.Records[0].Links {
		matches := collections.NewMap(twitterRe.SubexpNames(), twitterRe.FindStringSubmatch(link))
		newRequest := &rpb.Request{
			Id: rs.Id,
		}

		if match, ok := matches["twitter_id"]; ok && match != "" {
			newRequest.Target = &rpb.Source{
				ChannelId: matches["twitter_id"],
				Type:      rpb.SourceType_TWITTER,
			}
		} else {
			newRequest.Target = &rpb.Source{
				Url:  link,
				Type: rpb.SourceType_WEB,
			}
		}
		result = append(result, newRequest)
	}
	return result
}

func main() {
	flag.Parse()

	cfg := GetConfig()
	stg := storage.NewStorage(*cfg.StorageRoot, nil)
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
				response := chr.GetResponse()
				if response == nil {
					log.Warningf("Empty responses from %s", srcType)
					continue
				}

				recordSet := response.Result[0]
				request := response.Request
				for _, newRequest := range extractRequests(log, recordSet) {
					if targetChr, ok := adapters[newRequest.Target.Type]; ok {
						go targetChr.SubmitRequest(newRequest)
					}
				}
				err := stg.SaveRecords(recordSet)

				responseMessage := "Saved"
				if err != nil {
					responseMessage = "Error"
					log.Warningf("Error while saving a record: %s", err)
				}

				if request.Origin != nil && request.Origin.Type == rpb.SourceType_TELEGRAM {
					chr.SendMessage(&rpb.Message{
						Target: request.Origin, Content: []byte(responseMessage)})
				}
			}
		}(srcType, chr)
	}

	wg := sync.WaitGroup{}
	wg.Add(1)
	wg.Wait()
}
