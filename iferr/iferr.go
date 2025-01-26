package iferr

import (
	"fmt"
	"os"
)

func Exit[T any](value T, err error) T {
	if err != nil {
		fmt.Printf("Fatal error %s", err)
		os.Exit(-1)
	}
	return value
}
