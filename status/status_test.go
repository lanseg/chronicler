package status

import (
	"fmt"
	"reflect"
	"testing"

	conc "github.com/lanseg/golang-commons/concurrent"
	opt "github.com/lanseg/golang-commons/optional"

	sp "chronicler/status/status_go_proto"
)

const (
	testAddr = "localhost:12345"
)

type testBed struct {
	server   *statusServer
	client   *StatusClient
	tearDown func(tb testing.TB)
}

func setupServer(tb testing.TB) (*testBed, error) {
	server := NewStatusServer(testAddr)
	if err := server.Start(); err != nil {
		return nil, err
	}

	client, err := conc.WaitForSomething(func() opt.Optional[*StatusClient] {
		return opt.OfError(NewStatusClient(testAddr))
	}).Get()
	if err != nil {
		return nil, err
	}

	client.Start()

	return &testBed{
		client: client,
		server: server,
		tearDown: func(tb testing.TB) {
			server.Stop()
			client.Stop()
		},
	}, nil
}

func TestStatus(t *testing.T) {

	tb, err := setupServer(t)
	if err != nil {
		t.Errorf("Cannot initialize client and server: %s", err)
	}
	defer tb.tearDown(t)

	for _, tc := range []struct {
		name string
		put  []*sp.Metric
		want []*sp.Metric
	}{
		{
			name: "simple send receive",
			put:  []*sp.Metric{{Name: "metric", Value: &sp.Metric_IntValue{IntValue: int64(10)}}},
			want: []*sp.Metric{{Name: "metric", Value: &sp.Metric_IntValue{IntValue: int64(10)}}},
		},
		{
			name: "metric with the same name overwrites value",
			put: []*sp.Metric{
				{Name: "metric", Value: &sp.Metric_IntValue{IntValue: int64(10)}},
				{Name: "metric", Value: &sp.Metric_DoubleValue{DoubleValue: float64(10.20)}},
			},
			want: []*sp.Metric{
				{Name: "metric", Value: &sp.Metric_DoubleValue{DoubleValue: float64(10.20)}},
			},
		},
		{
			name: "metric with no value removes metric",
			put: []*sp.Metric{
				{Name: "metric", Value: &sp.Metric_IntValue{IntValue: int64(10)}},
				{Name: "metric", Value: &sp.Metric_DoubleValue{DoubleValue: float64(10.20)}},
				{Name: "metric"},
			},
			want: []*sp.Metric{},
		},
		{
			name: "all metric",
			put: []*sp.Metric{
				{Name: "metric", Value: &sp.Metric_IntValue{IntValue: int64(10)}},
				{Name: "metric1", Value: &sp.Metric_DoubleValue{DoubleValue: float64(10.20)}},
				{Name: "metric2", Value: &sp.Metric_StringValue{StringValue: "SomeValue"}},
				{Name: "metric3", Value: &sp.Metric_IntRangeValue{IntRangeValue: &sp.IntRange{
					MinValue: int64(0),
					MaxValue: int64(100500),
					Value:    int64(12345),
				}}},
				{Name: "metric4", Value: &sp.Metric_DoubleRangeValue{DoubleRangeValue: &sp.DoubleRange{
					MinValue: float64(-123.456),
					MaxValue: float64(123.456),
					Value:    float64(0.5),
				}}},
			},
			want: []*sp.Metric{
				{Name: "metric", Value: &sp.Metric_IntValue{IntValue: int64(10)}},
				{Name: "metric1", Value: &sp.Metric_DoubleValue{DoubleValue: float64(10.20)}},
				{Name: "metric2", Value: &sp.Metric_StringValue{StringValue: "SomeValue"}},
				{Name: "metric3", Value: &sp.Metric_IntRangeValue{IntRangeValue: &sp.IntRange{
					MinValue: int64(0),
					MaxValue: int64(100500),
					Value:    int64(12345),
				}}},
				{Name: "metric4", Value: &sp.Metric_DoubleRangeValue{DoubleRangeValue: &sp.DoubleRange{
					MinValue: float64(-123.456),
					MaxValue: float64(123.456),
					Value:    float64(0.5),
				}}},
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			tb.server.metrics = map[string]*sp.Metric{}
			for _, v := range tc.put {
				tb.client.PutValue(v)
			}

			if _, err := conc.WaitForSomething(func() opt.Optional[[]*sp.Metric] {
				vals, err := tb.client.GetValues()
				if err != nil {
					return opt.OfError[[]*sp.Metric](nil, err)
				}
				if !reflect.DeepEqual(tc.want, vals) {
					return opt.OfError[[]*sp.Metric](
						nil, fmt.Errorf("Expected to get %v, but got %v", tc.want, vals))
				}
				return opt.Of(vals)
			}).Get(); err != nil {
				t.Errorf("Could not read metrics from server: %s", err)
				return
			}
		})
	}

}
