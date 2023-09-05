package adapter

import (
	"chronicler/util"
	"sort"
	"time"

	"chronicler/records"
	rpb "chronicler/records/proto"
	"chronicler/twitter"

	"github.com/lanseg/golang-commons/collections"
)

type twitterRecordSource struct {
	RecordSource

	logger *util.Logger
	client twitter.Client
}

func NewTwitterAdapter(client twitter.Client) Adapter {
	tss := &twitterRecordSource{
		logger: util.NewLogger("TwitterAdapter"),
		client: client,
	}
	return NewAdapter("TwitterAdapter", tss, nil, false)
}

func (t *twitterRecordSource) GetRequestedRecords(request *rpb.Request) []*rpb.RecordSet {
	t.logger.Debugf("Got new request: %s", request)
	threadId := request.Target.ChannelId
	if threadId == "" {
		threadId = request.Target.MessageId
	}
	conv, _ := t.client.GetConversation(threadId)
	result := t.tweetToRecord(conv)
	result.Request = request
	return []*rpb.RecordSet{result}
}

func (t *twitterRecordSource) tweetToRecord(response *twitter.Response[twitter.Tweet]) *rpb.RecordSet {
	seen := collections.NewSet[string]([]string{})
	tweets := []twitter.Tweet{}
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

	recordsById := map[string]*rpb.Record{}
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
				t.logger.Warningf("Still missing media for key: %s", mediaKey)
				continue
			}
			twRecord.Files = append(twRecord.Files, &rpb.File{
				FileId:  m.Id,
				FileUrl: m.Url,
			})
		}
		recordsById[tweet.Id] = twRecord
	}

	um := []*rpb.UserMetadata{}
	rs := []*rpb.Record{}
	for _, user := range response.Includes.Users {
		um = append(um, &rpb.UserMetadata{
			Id:       user.Id,
			Username: user.Username,
			Quotes:   []string{user.Name},
		})
	}
	for _, tweet := range tweets {
		for _, ref := range tweet.Reference {
			if refTweet, ok := recordsById[ref.Id]; ok {
				recordsById[tweet.Id].Parent = recordsById[refTweet.Source.MessageId].Source
			}
		}
		rs = append(rs, recordsById[tweet.Id])
	}

	sort.Slice(rs, func(i int, j int) bool {
		return rs[i].Time < rs[j].Time
	})
	return records.NewRecordSet(rs, um)
}
