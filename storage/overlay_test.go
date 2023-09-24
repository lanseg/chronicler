package storage

import (
	"testing"
)

func staticId(id string) IdSource {
	return func() string {
		return id
	}
}

func TestOverlay(t *testing.T) {

	for _, tc := range []struct {
		name string
	}{
		{
			name: "Do something",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			o := NewOverlay(t.TempDir(), staticId("SomeId"))
			if o == nil {
				t.Errorf("Cannot create overlay")
			}
		})
	}
}
