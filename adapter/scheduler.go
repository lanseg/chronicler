package adapter

import (
	"time"

	cm "github.com/lanseg/golang-commons/common"
	conc "github.com/lanseg/golang-commons/concurrent"

	rpb "chronicler/records/proto"
)

func ScheduleRepeatedSource(provider SourceProvider, config *rpb.RequestConfig, dst chan<- *rpb.Request, duration time.Duration) {
	conc.RunPeriodically(func() {
		for _, src := range provider.GetSources() {
			dst <- &rpb.Request{
				Id:     cm.UUID4(),
				Config: config,
				Target: src,
			}
		}
	}, nil, duration)
}
