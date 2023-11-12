package adapter

import (
	rpb "chronicler/records/proto"
	"net/url"

	cm "github.com/lanseg/golang-commons/common"
)

type PikabuAdapter struct {
	Adapter

	logger *cm.Logger
}

func (p *PikabuAdapter) MatchLink(link string) *rpb.Source {
	if link == "" {
		return nil
	}
	u, err := url.Parse(link)
	if err != nil {
		return nil
	}
	return &rpb.Source{
		Url:  u.String(),
		Type: rpb.SourceType_PIKABU,
	}
}

func (p *PikabuAdapter) GetResponse(rq *rpb.Request) []*rpb.Response {
	result := []*rpb.Response{}
	return result
}

func (p *PikabuAdapter) SendMessage(m *rpb.Message) {
}
