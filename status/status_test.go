package status

import (
	"testing"

	conc "github.com/lanseg/golang-commons/concurrent"
	opt "github.com/lanseg/golang-commons/optional"
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

}
