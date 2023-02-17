package utils

func MapArray[T any, O any](in []T, f func(T) O) []O {
	out := make([]O, len(in))
	for i, elem := range in {
		out[i] = f(elem)
	}

	return out
}
