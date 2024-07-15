package tiktok

import (
	cm "github.com/lanseg/golang-commons/common"

	"chronicler/adapter"
	rpb "chronicler/records/proto"
	"regexp"
)

const (
	tiktokRe = ""
)

type tiktokAdapter struct {
	adapter.Adapter

	linkMatcher *regexp.Regexp
	logger      *cm.Logger
}

func NewTiktokAdapter() adapter.Adapter {
	return &tiktokAdapter{
		linkMatcher: regexp.MustCompile(tiktokRe),
		logger:      cm.NewLogger("TiktokAdapter"),
	}
}

func (t *tiktokAdapter) FindSources(r *rpb.Record) []*rpb.Source {
	return []*rpb.Source{}
}

func (t *tiktokAdapter) SendMessage(*rpb.Message) {
	t.logger.Warningf("TiktokAdapter cannot send messages")
}

func (t *tiktokAdapter) GetResponse(request *rpb.Request) []*rpb.Response {
	return []*rpb.Response{}
}
