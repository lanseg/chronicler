package adapter

import (
	"regexp"
	"sort"
	"time"

	rpb "chronicler/records/proto"
	"chronicler/twitter"

	"github.com/lanseg/golang-commons/collections"
	cm "github.com/lanseg/golang-commons/common"
)

const (
	twitterRe = "(\\.|^|//)(twitter|x.com).*/(?P<twitter_id>[0-9]+)[/]?"
)

type twitterAdapter struct {
	Adapter

	linkMatcher *regexp.Regexp
	logger      *cm.Logger
	client      twitter.Client
}

func NewTwitterAdapter(client twitter.Client) Adapter {
	return &twitterAdapter{
		linkMatcher: regexp.MustCompile(twitterRe),
		logger:      cm.NewLogger("TwitterAdapter"),
		client:      client,
	}
}

func (t *twitterAdapter) MatchLink(link string) *rpb.Source {
	matches := collections.NewMap(t.linkMatcher.SubexpNames(), t.linkMatcher.FindStringSubmatch(link))
	if match, ok := matches["twitter_id"]; ok && match != "" {
		return &rpb.Source{
			ChannelId: matches["twitter_id"],
			Type:      rpb.SourceType_TWITTER,
		}
	}
	return nil
}

func (t *twitterAdapter) SendMessage(*rpb.Message) {
	t.logger.Warningf("TwitterAdapter cannot send messages")
}

func (t *twitterAdapter) GetResponse(request *rpb.Request) []*rpb.Response {
	t.logger.Debugf("Got new request: %s", request)
	threadId := request.Target.ChannelId
	if threadId == "" {
		threadId = request.Target.MessageId
	}
	conv, _ := t.client.GetConversation(threadId)
	result := t.tweetToRecord(conv)
	result.Id = request.Id
	return []*rpb.Response{{
		Request: request,
		Result:  []*rpb.RecordSet{result},
	}}
}

func (t *twitterAdapter) tweetToRecord(response *twitter.Response[twitter.Tweet]) *rpb.RecordSet {
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
	return &rpb.RecordSet{
		Records:      rs,
		UserMetadata: um,
	}
}
