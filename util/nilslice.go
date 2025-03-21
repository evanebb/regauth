package util

// NilSliceToEmpty will convert a nil slice to an empty slice.
// If the slice is not nil, it will return the original slice.
// Mostly useful before encoding the slice as JSON, to ensure that an empty array is returned instead of null.
func NilSliceToEmpty[T any](s []T) []T {
	if s == nil {
		return make([]T, 0)
	}

	return s
}
