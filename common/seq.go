package common

import "iter"

func Map[S any, T any](data iter.Seq[S], f func(s S) T) iter.Seq[T] {
	return func(yield func(T) bool) {
		for obj := range data {
			if !yield(f(obj)) {
				return
			}
		}
	}
}

func FlatMap[S any, T any](data iter.Seq[S], f func(s S) iter.Seq[T]) iter.Seq[T] {
	return func(yield func(T) bool) {
		for obj := range data {
			for res := range f(obj) {
				if !yield(res) {
					return
				}
			}
		}
	}
}
