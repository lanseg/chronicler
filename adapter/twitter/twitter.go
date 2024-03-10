package twitter

import (
	"fmt"
	"regexp"
	"sort"
	"time"

	"chronicler/adapter"
	rpb "chronicler/records/proto"

	"github.com/lanseg/golang-commons/collections"
	cm "github.com/lanseg/golang-commons/common"
)

const (
	twitterRe = "(\\.|^|//)(twitter|x.com).*/(?P<twitter_id>[0-9]+)[/]?"
)

type twitterAdapter struct {
	adapter.Adapter

	linkMatcher *regexp.Regexp
	logger      *cm.Logger
	client      Client
}

func NewTwitterAdapter(client Client) adapter.Adapter {
	return &twitterAdapter{
		linkMatcher: regexp.MustCompile(twitterRe),
		logger:      cm.NewLogger("TwitterAdapter"),
		client:      client,
	}
}

func (t *twitterAdapter) FindSources(r *rpb.Record) []*rpb.Source {
	result := []*rpb.Source{}
	for _, link := range r.Links {
		if src := t.matchLink(link); src != nil {
			result = append(result, src)
		}
	}
	return result
}

func (t *twitterAdapter) matchLink(link string) *rpb.Source {
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
	if conv.Meta.ResultCount == 0 {
		return []*rpb.Response{}
	}
	result := t.tweetToRecord(conv)
	result.Id = cm.UUID4For(request.Target)
	return []*rpb.Response{{
		Request: request,
		Result:  []*rpb.RecordSet{result},
	}}
}

func (t *twitterAdapter) tweetToRecord(response *Response[Tweet]) *rpb.RecordSet {
	seen := collections.NewSet[string]([]string{})
	tweets := []Tweet{}
	for _, twt := range append(response.Data, response.Includes.Tweets...) {
		if seen.Contains(twt.Id) {
			continue
		}
		seen.Add(twt.Id)
		tweets = append(tweets, twt)
	}

	media := map[string]*TwitterMedia{}
	for _, m := range response.Includes.Media {
		bestMedia := GetBestQualityMedia(m)
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
