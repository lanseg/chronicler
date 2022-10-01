package util

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
