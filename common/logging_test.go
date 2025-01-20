package common

import (
	"testing"
)

func TestLogging(t *testing.T) {

	t.Run("Simple test", func(t *testing.T) {
		log := NewLogger("a logger")
		format := "Sample log arg1 %d arg2 %s"
		args := [](interface{}){1, "arg 2"}

		log.Debugf(format, args...)
		log.Infof(format, args...)
		log.Warningf(format, args...)
		log.Errorf(format, args...)
	})
}
