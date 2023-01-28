package util

func Ifnull[T any](value *T, ifnull *T) *T {
	if value == nil {
		return ifnull
	}
	return value
}
