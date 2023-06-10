package utils

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
