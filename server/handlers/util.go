package handlers

// convertSlice will convert a slice of type []I to a slice of type []O, using the provided conversion function to
// convert each item in the input slice.
func convertSlice[I any, O any](in []I, conversionFunc func(in I) O) []O {
	out := make([]O, len(in))
	for i := 0; i < len(in); i++ {
		out[i] = conversionFunc(in[i])
	}

	return out
}
