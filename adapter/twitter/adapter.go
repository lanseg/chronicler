package twitter

import (
	"fmt"
	"regexp"
	"sort"
	"time"

	"chronicler/adapter"
	"chronicler/common"
	opb "chronicler/proto"
)

var (
	twitterRe = regexp.MustCompile(`(\.|^|//)(twitter|x.com).*/(?P<twitter_id>[0-9]+)[/]?`)
)

func extractId(link string) string {
	subexp := twitterRe.SubexpNames()
	submatch := twitterRe.FindStringSubmatch(link)
	for i, key := range submatch {
		if subexp[i] == "twitter_id" {
			return key
		}
	}
	return ""
}

func getMissingTweetIds(knownTweets map[string]*Tweet, knownMedias map[string]*Media) []string {
	result := map[string]bool{}
	for _, tweet := range knownTweets {
		for _, ref := range tweet.Reference {
			if _, ok := knownTweets[ref.Id]; !ok {
				result[ref.Id] = true
			}
		}
		for _, key := range tweet.Attachments.MediaKeys {
			if _, ok := knownMedias[key]; !ok {
				result[tweet.Id] = true
			}
		}
	}
	keys := []string{}
	for t := range result {
		keys = append(keys, t)
	}
	return keys
}

func tweetKey(t *Tweet) string {
	return t.Id
}

func mediaKey(m *Media) string {
	return m.MediaKey
}

func userKey(u *User) string {
	return u.Id
}

func appendAll[T any](m map[string]*T, toAdd []*T, key func(*T) string) {
	for _, t := range toAdd {
		m[key(t)] = t
	}
}

type twitterAdapter struct {
	adapter.Adapter

	logger *common.Logger
	client Client
}

func NewAdapter(client Client) adapter.Adapter {
	return &twitterAdapter{
		logger: common.NewLogger("TwitterAdapter"),
		client: client,
	}
}

func (ta *twitterAdapter) Match(link *opb.Link) bool {
	return extractId(link.Href) != ""
}

func (ta *twitterAdapter) Get(link *opb.Link) ([]*opb.Object, error) {
	threadId := extractId(link.Href)
	if threadId == "" {
		return nil, fmt.Errorf("cannot extract thread id from link %q", link)
	}
	// Conversation
	allTweets := map[string]*Tweet{}
	allMedia := map[string]*Media{}
	allUsers := map[string]*User{}
	allConvs, err := ta.client.GetConversation(threadId)
	if err != nil {
		return nil, fmt.Errorf("cannot get conversation %q: %s", threadId, err)
	}
	for _, conv := range allConvs {
		appendAll(allTweets, conv.Data, tweetKey)
		appendAll(allTweets, conv.Includes.Tweets, tweetKey)
		appendAll(allMedia, conv.Includes.Media, mediaKey)
		appendAll(allUsers, conv.Includes.Users, userKey)
	}

	// Referenced Tweets
	ta.logger.Warningf("Still missing: %s", getMissingTweetIds(allTweets, allMedia))
	missingTweets, err := ta.client.GetTweets(getMissingTweetIds(allTweets, allMedia))
	if err != nil {
		return nil, fmt.Errorf("cannot get conversation %q: %s", threadId, err)
	}
	appendAll(allTweets, missingTweets.Data, tweetKey)
	appendAll(allTweets, missingTweets.Includes.Tweets, tweetKey)
	appendAll(allMedia, missingTweets.Includes.Media, mediaKey)
	appendAll(allUsers, missingTweets.Includes.Users, userKey)
	ta.logger.Warningf("Still missing: %s", getMissingTweetIds(allTweets, allMedia))

	return ta.tweetToObject(allTweets, allMedia, allUsers), nil
}

func (ta *twitterAdapter) tweetToObject(
	tweets map[string]*Tweet, medias map[string]*Media, users map[string]*User) []*opb.Object {
	result := []*opb.Object{}
	for tweetId, t := range tweets {
		obj := &opb.Object{
			Id:      tweetId,
			Content: []*opb.Content{{Text: t.Text, Mime: "text/plain"}},
		}

		generator := &opb.Generator{Id: t.AuthorId}
		if user, ok := users[t.AuthorId]; ok {
			generator.Name = user.Username
		}
		obj.Generator = append(obj.Generator, generator)

		if timestamp, err := time.Parse(time.RFC3339, t.Created); err == nil {
			obj.CreatedAt = &opb.Timestamp{
				Seconds: timestamp.Unix(),
			}
		}

		allKeys := map[string]bool{}
		for _, key := range t.Attachments.MediaKeys {
			allKeys[key] = true
		}
		for _, url := range t.Entities.Urls {
			if url.MediaKey != "" {
				allKeys[url.MediaKey] = true
			}
		}

		missingMediaKeys := map[string]bool{}
		for key := range allKeys {
			media, ok := medias[key]
			if !ok {
				missingMediaKeys[key] = true
				continue
			}
			br := int64(0)
			url := media.Url
			mediaType := ""
			for _, v := range media.Variants {
				if url == "" || v.Bitrate > br {
					url = v.Url
					br = v.Bitrate
					mediaType = v.ContentType
				}
			}
			if mediaType == "" {
				mediaType = common.GuessMimeType(url)
			}
			obj.Attachment = append(obj.Attachment, &opb.Attachment{
				Url:  url,
				Mime: mediaType,
			})
		}

		for _, url := range t.Entities.Urls {
			if _, ok := medias[url.MediaKey]; !ok {
				obj.Attachment = append(obj.Attachment, &opb.Attachment{
					Url:  url.ExpandedUrl,
					Mime: common.GuessMimeType(url.ExpandedUrl),
				})
			}
		}
		result = append(result, obj)
	}

	sort.Slice(result, func(i int, j int) bool {
		return result[i].CreatedAt.Seconds < result[j].CreatedAt.Seconds
	})
	return result
}
