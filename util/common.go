package util

type Pair[K any, V any] struct {
	First  K
	Second V
}

func AsPair[K any, V any](first K, second V) Pair[K, V] {
	return Pair[K, V]{
		First:  first,
		Second: second,
	}
}
