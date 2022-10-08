package util

import (
	"reflect"
	"testing"
)

func identity[T any](a T) T {
	return a
}

func TestGroupBy(t *testing.T) {
	for _, tc := range []struct {
		desc     string
		items    []string
		key      func(a string) string
		expected map[string]([]string)
	}{
		{
			desc:  "successful operation for a simple array",
			items: []string{"hello world", "hello there", "goodbye world"},
			key: func(a string) string {
				return string(a[0])
			},
			expected: map[string]([]string){
				"h": []string{"hello world", "hello there"},
				"g": []string{"goodbye world"},
			},
		},
		{
			desc:     "successful operation for an empty array",
			items:    []string{},
			key:      identity[string],
			expected: map[string]([]string){},
		},
		{
			desc:  "successful operation for an identity function key",
			items: []string{"hello world", "hello there", "goodbye world"},
			key:   identity[string],
			expected: map[string]([]string){
				"hello world":   []string{"hello world"},
				"hello there":   []string{"hello there"},
				"goodbye world": []string{"goodbye world"},
			},
		},
		{
			desc:  "duplicates should stay in the result",
			items: []string{"ab", "ab", "ab", "ba", "ba", "ba"},
			key: func(a string) string {
				return string(a[0])
			},
			expected: map[string]([]string){
				"a": []string{"ab", "ab", "ab"},
				"b": []string{"ba", "ba", "ba"},
			},
		},
	} {
		t.Run(tc.desc, func(t *testing.T) {
			result := GroupBy(tc.items, tc.key)
			if !reflect.DeepEqual(result, tc.expected) {
				t.Errorf("GroupBy(%v) expected to be %v, but got %v", tc.items,
					tc.expected, result)
			}
		})
	}
}

func TestKeysValues(t *testing.T) {
	for _, tc := range []struct {
		desc       string
		data       map[string]string
		wantKeys   []string
		wantValues []string
	}{} {
		t.Run(tc.desc, func(t *testing.T) {
			keys := Keys(tc.data)
			values := Values(tc.data)
			if !reflect.DeepEqual(tc.wantKeys, keys) {
				t.Errorf("Keys(%v) expected to be %v, but got %v", tc.data, tc.wantKeys, keys)
			}
			if !reflect.DeepEqual(tc.wantValues, values) {
				t.Errorf("Values(%v) expected to be %v, but got %v", tc.data, tc.wantValues, values)
			}
		})
	}
}
