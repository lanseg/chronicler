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

func getWholeConversation(client twitter.Client, conversation string) *rpb.RecordSet {
	token := ""
	tweets := []*twitter.Tweet{}
	seen := util.NewSet[string]([]string{})
	for {
		result, err := client.GetConversation(conversation, token)
		if err != nil {
			log.Errorf("Cannot load tweet: %s", err)
			break
		}
		token = result.Meta.NextToken
		for _, t := range append(result.Data, result.Includes.Tweets...) {
			if seen.Contains(t.Id) {
				continue
			}
			tweets = append(tweets, t)
			seen.Add(t.Id)
		}
		if len(result.Data) == 0 || token == "" {
			break
		}
	}

	records := map[string]*rpb.Record{}
	for _, tweet := range tweets {
		twRecord := &rpb.Record{
			Source: &rpb.Source{
				SenderId:  tweet.Author,
				ChannelId: conversation,
				MessageId: tweet.Id,
				Type:      rpb.SourceType_TWITTER,
			},
			TextContent: tweet.Text,
		}
		if timestamp, err := time.Parse(time.RFC3339, tweet.Created); err == nil {
			twRecord.Time = timestamp.Unix()
		}
		for _, m := range tweet.Media {
			twRecord.Files = append(twRecord.Files, &rpb.File{
				FileId:  m.Id,
				FileUrl: m.Url,
			})
		}
		records[tweet.Id] = twRecord
	}

	result := &rpb.RecordSet{}
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

func main() {
	flag.Parse()

	twt := twitter.NewClient(*twitterApiKey)
	stg := storage.NewStorage(*storageRoot)

	for _, arg := range flag.Args() {
		threadId := arg
		if err := stg.SaveRecords(threadId, getWholeConversation(twt, threadId)); err != nil {
			log.Warningf("Error while saving a record: %s", err)
		}
	}
}
