package util

type Set[T comparable] struct {
  amap map[T]bool
}

func (s *Set[T]) put(value T) bool {
  _, ok := s.amap[value]
  s.amap[value] = true
  return ok
}

func GroupBy[T, V comparable] (items []T, key func(a T) V) map[V]([]T) {
  result := map[V]([]T){}
  for _, item := range items {
    k := key(item)
    result[k] = append(result[k], item)
  }
  return result
}

func Values[K comparable, V any] (m map[K]V) []V {
  result := []V{}
  for _, v := range m {
    result = append(result, v)
  }
  return result
}

func Keys[K comparable, V any] (m map[K]V) []K {
  result := []K{}
  for k, _ := range m {
    result = append(result, k)
  }
  return result  
}

