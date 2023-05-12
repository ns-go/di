package utils

func FilterSlice[T any](sl []T, filterFunc func(T) bool) []T {
	result := make([]T, 0)
	for _, el := range sl {
		if filterFunc(el) {
			result = append(result, el)
		}
	}

	return result
}

func MapSlice[T any, TResult any](sl []T, mapFunc func(T) TResult) []TResult {
	result := make([]TResult, len(sl))
	for i := 0; i < len(sl); i++ {
		result[i] = mapFunc(sl[i])
	}

	return result
}

func Ptr[T any](val T) *T {
	return &val
}
