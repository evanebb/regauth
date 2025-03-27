package local

import (
	"errors"
	"strings"
	"testing"
)

func TestUserCredentials_SetPassword(t *testing.T) {
	t.Parallel()

	t.Run("weak password error", func(t *testing.T) {
		credentials := UserCredentials{}

		password := strings.Repeat("a", 8)
		if err := credentials.SetPassword(password); !errors.Is(err, ErrWeakPassword) {
			t.Errorf("expected %q, got %q", ErrWeakPassword, err)
		}
	})

	t.Run("successfully set password", func(t *testing.T) {
		credentials := UserCredentials{}

		password := strings.Repeat("a", 9)
		if err := credentials.SetPassword(password); err != nil {
			t.Errorf("expected error to be nil, got %q", err)
		}
	})
}

func TestUserCredentials_CheckPassword(t *testing.T) {
	t.Parallel()

	t.Run("valid password", func(t *testing.T) {
		credentials := UserCredentials{}

		password := strings.Repeat("a", 9)
		if err := credentials.SetPassword(password); err != nil {
			t.Fatalf("expected error to be nil, got %q", err)
		}

		if err := credentials.CheckPassword(password); err != nil {
			t.Fatalf("expected password to match, got %q", err)
		}
	})

	t.Run("invalid password", func(t *testing.T) {
		credentials := UserCredentials{}

		password := strings.Repeat("a", 9)
		if err := credentials.SetPassword(password); err != nil {
			t.Fatalf("expected error to be nil, got %q", err)
		}

		if err := credentials.CheckPassword("foobar"); err == nil {
			t.Fatalf("expected password to not match, got %q", err)
		}
	})
}
