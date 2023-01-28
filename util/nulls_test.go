package util

import (
	"testing"
)

func strptr(s string) *string {
	var a *string = &s
	return a
}

func TestNulls(t *testing.T) {
	t.Run("Both non null", func(t *testing.T) {
		a := strptr("a")
		b := strptr("b")
		if result := Ifnull(a, b); result != a {
			t.Errorf("Expected Ifnull(%v, %v) = %v, but got %v", a, b, a, result)
		}
	})

	t.Run("both null", func(t *testing.T) {
		var a *string = nil
		var b *string = nil
		if result := Ifnull(a, b); result != nil {
			t.Errorf("expected ifnull(%v, %v) = %v, but got %v", a, b, nil, result)
		}
	})

	t.Run("first is null", func(t *testing.T) {
		var a *string = nil
		b := strptr("b")
		if result := Ifnull(a, b); result != b {
			t.Errorf("expected ifnull(%v, %v) = %v, but got %v", a, b, b, result)
		}
	})
}
