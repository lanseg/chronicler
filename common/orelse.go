package common

import (
	"fmt"
	"os"
)

func OrExit[T any](value T, err error) T {
	if err != nil {
		fmt.Printf("Fatal error %s", err)
		os.Exit(-1)
	}
	return value
}
