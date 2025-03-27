package auth

import (
	"errors"
	"testing"
)

func TestTokenHashing(t *testing.T) {
	t.Parallel()

	raw := "foobarbaz"
	hash := HashTokenWithRandomSalt(raw)

	t.Run("matching token", func(t *testing.T) {
		if err := CompareTokenAndHash(raw, hash); err != nil {
			t.Errorf("expected token to match hash, got error %q", err)
		}
	})

	t.Run("non-matching token", func(t *testing.T) {
		if err := CompareTokenAndHash("barbazfoo", hash); !errors.Is(err, ErrHashDoesNotMatch) {
			t.Errorf("expected token to not match hash, got %q", err)
		}
	})

	t.Run("hex decode error", func(t *testing.T) {
		invalidHash := []byte("foo")
		if err := CompareTokenAndHash("", invalidHash); err == nil || errors.Is(err, ErrHashDoesNotMatch) {
			t.Errorf("expected hex decode error, got %q", err)
		}
	})
}
