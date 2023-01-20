package main

import (
	"flag"
	"sort"
	"time"

	"chronist/storage"
	"chronist/twitter"
	"chronist/util"

	rpb "chronist/proto/records"
)

const (
	twitterApiFlag  = "twitter_api_key"
	storageRootFlag = "storage_root"
)

var (
	twitterApiKey = flag.String(twitterApiFlag, "", "A key for the twitter api.")
	storageRoot   = flag.String(storageRootFlag, "chronist_storage", "A local folder to save downloads.")
	log           = util.NewLogger("main")
)

func twitterToRecord(response *twitter.Response) *rpb.RecordSet {
	seen := util.NewSet[string]([]string{})
	tweets := []*twitter.Tweet{}
	for _, twt := range append(response.Data, response.Includes.Tweets...) {
		if seen.Contains(twt.Id) {
			continue
		}
		seen.Add(twt.Id)
		tweets = append(tweets, twt)
	}

	media := map[string]*twitter.TwitterMedia{}
	for _, m := range response.Includes.Media {
		bestMedia := twitter.GetBestQualityMedia(m)
		media[bestMedia.Id] = bestMedia
	}

	records := map[string]*rpb.Record{}
	for _, tweet := range tweets {
		twRecord := &rpb.Record{
			Source: &rpb.Source{
				SenderId:  tweet.Author,
				ChannelId: tweet.ConversationId,
				MessageId: tweet.Id,
				Type:      rpb.SourceType_TWITTER,
			},
			TextContent: tweet.Text,
		}
		if timestamp, err := time.Parse(time.RFC3339, tweet.Created); err == nil {
			twRecord.Time = timestamp.Unix()
		}
		for _, mediaKey := range tweet.Attachments.MediaKeys {
			m, ok := media[mediaKey]
			if !ok {
				log.Warningf("Missing media for key: %s", mediaKey)
			}
			twRecord.Files = append(twRecord.Files, &rpb.File{
				FileId:  m.Id,
				FileUrl: m.Url,
			})
		}
		records[tweet.Id] = twRecord
	}

	result := &rpb.RecordSet{}
	for _, user := range response.Includes.Users {
		result.UserMetadata = append(result.UserMetadata, &rpb.UserMetadata{
			Id:       user.Id,
			Username: user.Username,
			Quotes:   []string{user.Name},
		})
	}
	for _, tweet := range tweets {
		for _, ref := range tweet.Reference {
			if refTweet, ok := records[ref.Id]; ok {
				records[refTweet.Source.MessageId].Parent = records[tweet.Id].Source
			}
		}
		result.Records = append(result.Records, records[tweet.Id])
	}

	sort.Slice(result.Records, func(i int, j int) bool {
		return result.Records[i].Time < result.Records[j].Time
	})
	return result
}

func parseRequest(s string) rpb.Request {

	return rpb.Request{
		RawRequest: s,
		Source: &rpb.Source{
			ChannelId: s,
			Type:      rpb.SourceType_TWITTER,
		},
	}
}

func main() {
	flag.Parse()

	twt := twitter.NewClient(*twitterApiKey)
	stg := storage.NewStorage(*storageRoot)

	for _, arg := range flag.Args() {
		request := parseRequest(arg)
		switch srcType := request.Source.Type; srcType {
		case rpb.SourceType_TWITTER:
			threadId := request.Source.ChannelId
			conv, err := twt.GetConversation(threadId)
			if err != nil {
				log.Errorf("Failed to get conversation for id %s: %s", threadId, err)
			}
			if err := stg.SaveRecords(threadId, twitterToRecord(conv)); err != nil {
				log.Warningf("Error while saving a record: %s", err)
			}
		default:
			log.Warningf("No loader found for request %s", request)
		}
	}
}
