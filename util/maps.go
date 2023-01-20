package util

func Identity[T any](value T) T {
	return value
}

func Unique[T comparable](values []T) []T {
	return Keys(GroupBy(values, Identity[T]))
}

func GroupBy[T any, V comparable](items []T, key func(a T) V) map[V]([]T) {
	result := map[V]([]T){}

	for _, item := range items {
		k := key(item)
		result[k] = append(result[k], item)
	}

	return result
}

func Values[K comparable, V any](m map[K]V) []V {
	result := []V{}

	for _, v := range m {
		result = append(result, v)
	}

	return result
}

func Keys[K comparable, V any](m map[K]V) []K {
	result := []K{}

	for k := range m {
		result = append(result, k)
	}

	return result
}

func Max[T any](values []T, comp func(a T) int) T {
	maxValue := 0
	maxIndex := 0
	for i, v := range values {
		val := comp(v)
		if (i == 0) || (i > 0 && val > maxValue) {
			maxValue = val
			maxIndex = i
		}
	}
	return values[maxIndex]
}

type Set[T comparable] struct {
	values map[T]bool
}

func NewSet[T comparable](values []T) *Set[T] {
	setValues := map[T]bool{}
	for _, v := range values {
		setValues[v] = true
	}
	return &Set[T]{
		values: setValues,
	}
}

func (s *Set[T]) Size() int {
	return len(s.values)
}

func (s *Set[T]) Values() []T {
	return Keys(s.values)
}

func (s *Set[T]) Add(item T) {
	s.values[item] = true
}

func (s *Set[T]) AddSet(items *Set[T]) {
	for item := range items.values {
		s.values[item] = true
	}
}

func (s *Set[T]) AddAll(items []T) {
	for _, item := range items {
		s.values[item] = true
	}
}

func (s *Set[T]) Remove(item T) {
	if _, ok := s.values[item]; ok {
		delete(s.values, item)
	}
}

func (s *Set[T]) Contains(item T) bool {
	_, ok := s.values[item]
	return ok
}

func (s *Set[T]) Clear() {
	s.values = map[T]bool{}
}

func NewMap[K comparable, V any](keys []K, values []V) map[K]V {
	result := make(map[K]V, len(keys))
	for i, k := range keys {
		var value V
		if i < len(values) {
			value = values[i]
		}
		result[k] = value
	}
	return result
}
