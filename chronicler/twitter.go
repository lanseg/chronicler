package chronicler

import (
	"chronicler/util"
	"sort"
	"time"

	rpb "chronicler/proto/records"
	"chronicler/twitter"

	"github.com/lanseg/golang-commons/collections"
)

type Twitter struct {
	Chronicler

	requests chan *rpb.Request
	records  chan *rpb.RecordSet
	logger   *util.Logger
	client   twitter.Client
}

func NewTwitterChronicler(client twitter.Client) Chronicler {
	twitterChronicler := &Twitter{
		logger:   util.NewLogger("Twitter"),
		records:  make(chan *rpb.RecordSet),
		requests: make(chan *rpb.Request),
		client:   client,
	}
	go twitterChronicler.requestLoop()
	return twitterChronicler
}

func (t *Twitter) GetRecordSource() <-chan *rpb.RecordSet {
	return t.records
}

func (t *Twitter) SubmitRequest(r *rpb.Request) {
	t.requests <- r
}

func (t *Twitter) SendResponse(*rpb.Response) {
	t.logger.Debugf("SendResponse doesn't work for Twitter by design")
}

func (t *Twitter) requestLoop() {
	t.logger.Debugf("Starting request loop")
	for {
		request := <-t.requests
		t.logger.Debugf("Got new request: %s", request)
		threadId := request.Source.ChannelId
		if threadId == "" {
			threadId = request.Source.MessageId
		}
		conv, _ := t.client.GetConversation(threadId)
		t.records <- t.tweetToRecord(conv)
	}
}

func (t *Twitter) tweetToRecord(response *twitter.Response[twitter.Tweet]) *rpb.RecordSet {
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
				t.logger.Warningf("Still missing media for key: %s", mediaKey)
				continue
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
				records[tweet.Id].Parent = records[refTweet.Source.MessageId].Source
			}
		}
		result.Records = append(result.Records, records[tweet.Id])
	}

	sort.Slice(result.Records, func(i int, j int) bool {
		return result.Records[i].Time < result.Records[j].Time
	})
	return result
}
