package status

import (
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

	return &testBed{
		client: client,
		server: server,
		tearDown: func(tb testing.TB) {
			server.Stop()
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
	} {
		t.Run(tc.name, func(t *testing.T) {
			for _, v := range tc.put {
				if err := tb.client.PutValue(v); err != nil {
					t.Errorf("Could not send metrics to server: %s", err)
					return
				}
			}

			values, err := tb.client.GetValues()
			if err != nil {
				t.Errorf("Could not read metrics from server: %s", err)
				return
			}

			if !reflect.DeepEqual(tc.want, values) {
				t.Errorf("Expected metrics to be %v, but got %v", tc.want, values)
			}
		})
	}

}
