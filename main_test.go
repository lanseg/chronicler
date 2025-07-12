package main

import (
	"os"
	"testing"
)

func doTest(expectExit int, args ...string) func(t *testing.T) {
	return func(t *testing.T) {
		baseArgs := append([]string{}, os.Args...)
		os.Args = os.Args[:1]
		code := mainf()
		if code != -2 {
			t.Errorf("Expected exit code %d, but got %d", expectExit, code)
		}
		os.Args = baseArgs
	}
}

func TestMain(t *testing.T) {
	t.Run("main no arguments", doTest(-2))
	t.Run("main help argument", doTest(-2, "--help"))
}
