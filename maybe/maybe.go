package maybe

import (
	"errors"
)

var NoElementError = errors.New("Trying to get from Nothing")

type Func[A, B any] func(x A) B

type Predicate[T any] Func[T, bool]

func True[T any](_ T) bool {
	return true
}

func False[T any](_ T) bool {
	return false
}

type Maybe[T any] interface {
	IsPresent() bool
	Filter(p Predicate[T]) Maybe[T]
	OrElse(other T) T
	Get() (T, error)
}

type Just[T any] struct {
	Maybe[T]

	value T
}

func (j Just[T]) IsPresent() bool {
	return true
}

func (j Just[T]) Filter(p Predicate[T]) Maybe[T] {
	if p(j.value) {
		return j
	}
	return Nothing[T]{}
}

func (j Just[T]) OrElse(other T) T {
	return j.value
}

func (j Just[T]) Get() (T, error) {
	return j.value, nil
}

type Nothing[T any] struct {
	Maybe[T]
}

func (n Nothing[T]) IsPresent() bool {
	return false
}

func (n Nothing[T]) Filter(p Predicate[T]) Maybe[T] {
	return n
}

func (n Nothing[T]) OrElse(other T) T {
	return other
}

func (n Nothing[T]) Get() (T, error) {
	return *new(T), NoElementError
}

type Error[T any] struct {
	Nothing[T]

	err error
}

func (e Error[T]) Get() (T, error) {
	return *new(T), e.err
}

func Map[U, V any](m Maybe[U], f Func[U, V]) Maybe[V] {
	switch m.(type) {
	case Just[U]:
		return Just[V]{
			value: f(m.(Just[U]).value),
		}
	}
	return Nothing[V]{}
}

func FlatMap[U, V any](m Maybe[U], f Func[U, Maybe[V]]) Maybe[V] {
	switch m.(type) {
	case Just[U]:
		return f(m.(Just[U]).value)
	}
	return Nothing[V]{}
}

func Return[T any](x T) Maybe[T] {
	return Just[T]{value: x}
}

func OfNullable[T any](x *T) Maybe[T] {
	if x == nil {
		return Nothing[T]{}
	}

	return Just[T]{value: *x}
}
