package utils

import "testing"

func TestOrderedMap_GetOrDefault(t *testing.T) {
	oMap := NewOMap[int, *[]string]()
	oMap.Upsert(5, func(value *[]string) {
		*value = append(*value, "new")
	}, newListString)
	oMap.Upsert(5, func(value *[]string) {
		*value = append(*value, "another")
	}, newListString)

	oMap.Upsert(3, func(value *[]string) {
		*value = append(*value, "another")
	}, newListString)

	oMap.Reverse(func(k int, v *[]string) {
		print(k)
		println(v)
	})
}

func newListString() *[]string {
	return new([]string)
}
