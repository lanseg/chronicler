package chronicler

import (
	"fmt"
	"time"

	cm "github.com/lanseg/golang-commons/common"
	conc "github.com/lanseg/golang-commons/concurrent"

	"chronicler/adapter"
	rpb "chronicler/records/proto"
	"chronicler/status"
)

func ScheduleRepeatedSource(stats status.StatusClient, name string, provider adapter.SourceProvider, engine rpb.WebEngine, ch Chronicler, duration time.Duration) {
	conc.RunPeriodically(func() {
		stats.PutDateTime(fmt.Sprintf("%s.last_provide", name), time.Now())
		for _, src := range provider.GetSources() {
			ch.SubmitRequest(&rpb.Request{
				Id: cm.UUID4(),
				Config: &rpb.RequestConfig{
					Engine: engine,
				},
				Target: src,
			})
		}
	}, nil, duration)
}
