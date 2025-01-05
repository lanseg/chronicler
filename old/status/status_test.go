package status

import (
	"fmt"
	"reflect"
	"sort"
	"testing"
	"time"

	conc "github.com/lanseg/golang-commons/concurrent"
	opt "github.com/lanseg/golang-commons/optional"

	sp "chronicler/status/status_go_proto"
)

const (
	testAddr = "localhost:12345"
)

type testBed struct {
	server   *statusServer
	client   StatusClient
	tearDown func(tb testing.TB)
}

func (tb *testBed) WaitForValues(want []*sp.Metric) opt.Optional[[]*sp.Metric] {
	return conc.WaitForSomething(func() opt.Optional[[]*sp.Metric] {
		vals, err := tb.client.GetValues()
		if err != nil {
			return opt.OfError[[]*sp.Metric](nil, err)
		}
		sort.Slice(vals, func(i int, j int) bool {
			return vals[i].Name < vals[j].Name
		})
		sort.Slice(want, func(i int, j int) bool {
			return want[i].Name < want[j].Name
		})
		if !reflect.DeepEqual(want, vals) {
			return opt.OfError[[]*sp.Metric](
				nil, fmt.Errorf("Expected to get %v, but got %v", want, vals))
		}
		return opt.Of(vals)
	})
}

func setupServer(tb testing.TB) (*testBed, error) {
	server := NewStatusServer(testAddr)
	if err := server.Start(); err != nil {
		return nil, err
	}

	client, err := conc.WaitForSomething(func() opt.Optional[StatusClient] {
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
			client.Stop()
			server.Stop()
		},
	}, nil
}

func TestStatusPutShortcuts(t *testing.T) {
	tb, err := setupServer(t)
	if err != nil {
		t.Errorf("Cannot initialize client and server: %s", err)
	}
	defer tb.tearDown(t)

	tb.client.PutInt("int", 10)
	tb.client.PutDouble("double", 3.1415)
	tb.client.PutString("string", "string")
	tb.client.PutIntRange("intrange", 10, -10, 100)
	tb.client.PutDoubleRange("doublerange", 1.234, -0.5, 11.11)
	sometime, _ := time.Parse(time.RFC3339, "2024-01-02T15:04:05+07:00")
	tb.client.PutDateTime("datetime", sometime)

	want := []*sp.Metric{
		{Name: "int", Value: &sp.Metric_IntValue{IntValue: int64(10)}},
		{Name: "double", Value: &sp.Metric_DoubleValue{DoubleValue: float64(3.1415)}},
		{Name: "string", Value: &sp.Metric_StringValue{StringValue: "string"}},
		{Name: "intrange", Value: &sp.Metric_IntRangeValue{IntRangeValue: &sp.IntRange{
			Value: int64(10), MinValue: int64(-10), MaxValue: int64(100)}}},
		{Name: "doublerange", Value: &sp.Metric_DoubleRangeValue{DoubleRangeValue: &sp.DoubleRange{
			Value: float64(1.234), MinValue: float64(-0.5), MaxValue: float64(11.11)}}},
		{Name: "datetime", Value: &sp.Metric_DateTimeValue{DateTimeValue: &sp.DateTime{
			Timestamp: int64(1704182645000), Offset: int64(25200)}}},
	}

	if _, err := tb.WaitForValues(want).Get(); err != nil {
		t.Errorf("Could not read metrics from server: %s", err)
		return
	}

	want = []*sp.Metric{
		{Name: "int", Value: &sp.Metric_IntValue{IntValue: int64(10)}},
		{Name: "doublerange", Value: &sp.Metric_DoubleRangeValue{DoubleRangeValue: &sp.DoubleRange{
			Value: float64(1.234), MinValue: float64(-0.5), MaxValue: float64(11.11)}}},
	}
	tb.client.DeleteMetric("double")
	tb.client.DeleteMetric("string")
	tb.client.DeleteMetric("intrange")
	tb.client.DeleteMetric("datetime")

	if _, err := tb.WaitForValues(want).Get(); err != nil {
		t.Errorf("Could not read metrics from server: %s", err)
		return
	}
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
				{Name: "datetime", Value: &sp.Metric_DateTimeValue{DateTimeValue: &sp.DateTime{
					Timestamp: int64(10005000), Offset: int64(123)}}},
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
				{Name: "datetime", Value: &sp.Metric_DateTimeValue{DateTimeValue: &sp.DateTime{
					Timestamp: int64(10005000), Offset: int64(123)}}},
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			tb.server.metrics = map[string]*sp.Metric{}
			for _, v := range tc.put {
				tb.client.PutValue(v)
			}

			if _, err := tb.WaitForValues(tc.want).Get(); err != nil {
				t.Errorf("Could not read metrics from server: %s", err)
				return
			}
		})
	}

}
