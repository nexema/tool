package utils

func MapArray[T any, O any](in []T, f func(T) O) []O {
	out := make([]O, len(in))
	for i, elem := range in {
		out[i] = f(elem)
	}

	return out
}

func Find[T any](arr *[]T, predicate func(*T) bool) *T {
	for _, elem := range *arr {
		if predicate(&elem) {
			return &elem
		}
	}

	return nil
}

func Contains[T comparable](arr *[]T, elem T) bool {
	for _, item := range *arr {
		if item == elem {
			return true
		}
	}

	return false
}
