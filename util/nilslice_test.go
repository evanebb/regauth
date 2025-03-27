package util

import "testing"

func TestNilSliceToEmpty(t *testing.T) {
	t.Parallel()

	t.Run("nil slice is converted to empty slice", func(t *testing.T) {
		t.Parallel()
		var s []string
		if actual := NilSliceToEmpty(s); actual == nil {
			t.Fatalf("expected slice to not be nil")
		}
	})

	t.Run("non-nil slice is returned as-is", func(t *testing.T) {
		t.Parallel()
		s := make([]string, 0)
		actual := NilSliceToEmpty(s)
		// not really anything to check here, just check the length I guess
		if len(actual) != 0 || actual == nil {
			t.Fatalf("expected slice to not be changed")
		}
	})
}
