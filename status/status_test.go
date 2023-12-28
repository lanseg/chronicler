package status

import (
	"testing"
)

func TestNoopStatus(t *testing.T) {
	nos := NewNoopStatus()
	nj := nos.NewJob("Hello job")
	nj.Start()
	nj.Stop()
	nj.Progress()
	nj.GetState()
}
