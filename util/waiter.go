package util

import (
	"fmt"
	"time"

	"github.com/lanseg/golang-commons/optional"
)

type waiter[T any] struct {
	lastResult    optional.Optional[T]
	valueProvider func() optional.Optional[T]
}

func (w *waiter[T]) Get() optional.Optional[T] {
	w.lastResult = w.valueProvider()
	return w.lastResult
}

func WaitFor(condition func() (bool, error), retries int, interval time.Duration) error {
	for ; retries > 0; retries-- {
		done, err := condition()
		if done && err == nil {
			return nil
		}
		if err != nil {
			return fmt.Errorf("Failed to reach the condition because of error %s", err)
		}
		time.Sleep(interval)
	}
	return fmt.Errorf("Failed to reach the condition after retrying for %d times", retries)
}

func WaitForPresent[T any](provider func() optional.Optional[T], retries int, interval time.Duration) optional.Optional[T] {
	w := &waiter[T]{
		valueProvider: provider,
	}
	WaitFor(func() (bool, error) {
		return w.Get().IsPresent(), nil
	}, retries, interval)
	return w.lastResult
}
