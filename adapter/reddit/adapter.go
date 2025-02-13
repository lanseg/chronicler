package reddit

import (
	"chronicler/adapter"
	"chronicler/common"
	opb "chronicler/proto"
	"regexp"
)

// (\.|^|//)(twitter|x.com).*/(?P<twitter_id>[0-9]+)[/]?`
// https://www.reddit.com/r/law/comments/1inaszr/comment/mc9tny7/?utm_source=share&utm_medium=web3x&utm_name=web3xcss&utm_term=1&utm_content=share_button
var (
	redditRe = regexp.MustCompile(`reddit.com/r/(?P<subreddit>[^/]*)/comments/(?P<author>[^/]*)/(?P<maybePostName>[^/$]*)(?:/(?P<maybeCommentId>[^/$]*))?`)
)

type redditAdapter struct {
	adapter.Adapter

	logger *common.Logger
	client adapter.HttpClient
}

func NewAdapter(client adapter.HttpClient) adapter.Adapter {
	return &redditAdapter{
		logger: common.NewLogger("TwitterAdapter"),
		client: client,
	}
}

func (ta *redditAdapter) Match(link *opb.Link) bool {
	postDef := ParseLink(link.Href)
	return postDef.Author != "" && postDef.MaybePost != "" && postDef.Subreddit != ""
}

func (ta *redditAdapter) Get(link *opb.Link) ([]*opb.Object, error) {
	result := []*opb.Object{}
	return result, nil
}
