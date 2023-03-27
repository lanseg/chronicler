package maybe

import (
	"testing"
)

func TestRunners(t *testing.T) {
	for _, tc := range []struct {
		desc string
	}{} {
		t.Run(tc.desc, func(t *testing.T) {
		})
	}
}

func TestJust(t *testing.T) {
	value := "whatever"
	just := Just[string]{
		value: value,
	}

	t.Run("IsPresent", func(t *testing.T) {
		if !just.IsPresent() {
			t.Errorf("Just must return true, but got false")
		}
	})

	t.Run("Get", func(t *testing.T) {
		if result, err := just.Get(); result != value || err != nil {
			t.Errorf("Just must return %s, but got %s, %s instead", value, result, err)
		}
	})

	t.Run("MatchesFilter", func(t *testing.T) {
		if result := just.Filter(True[string]); result != just {
			t.Errorf("Just must return itself (%s), but got %s instead", just, result)
		}
	})

	t.Run("DoesntMatchFilter", func(t *testing.T) {
		if result := just.Filter(False[string]); result == just {
			t.Errorf("Just must return Nothing (%s), but got %s instead", Nothing[string]{}, result)
		}
	})

	t.Run("OrElseReturnsItself", func(t *testing.T) {
		if result := just.OrElse("other"); result != value {
			t.Errorf("Just must return value (%s), but got %s instead", value, result)
		}
	})
}

func TestNothing(t *testing.T) {
	value := "whatever"
	nothing := Nothing[string]{}

	t.Run("IsPresent", func(t *testing.T) {
		if nothing.IsPresent() {
			t.Errorf("Nothing must return true, but got false")
		}
	})

	t.Run("Get", func(t *testing.T) {
		if result, err := nothing.Get(); result == value || err != NoElementError {
			t.Errorf("Nothing must return %s, %s but got %s, %s instead", nil, NoElementError, result, err)
		}
	})

	t.Run("MatchesFilter", func(t *testing.T) {
		if result := nothing.Filter(True[string]); result != nothing {
			t.Errorf("Nothing must return itself (%s), but got %s instead", nothing, result)
		}
	})

	t.Run("DoesntMatchFilter", func(t *testing.T) {
		if result := nothing.Filter(False[string]); result != nothing {
			t.Errorf("Nothing must return itself (%s), but got %s instead", nothing, result)
		}
	})

	t.Run("OrElseReturnsItself", func(t *testing.T) {
		if result := nothing.OrElse("other"); result != "other" {
			t.Errorf("Nothing must return value (%s), but got %s instead", "other", result)
		}
	})
}
