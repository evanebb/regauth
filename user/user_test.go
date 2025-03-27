package user

import (
	"errors"
	"strings"
	"testing"
)

func TestUser_IsValid(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		desc string
		user User
		err  error
	}{
		{"valid user", User{Username: Username("username"), Role: RoleAdmin}, nil},
		{"invalid username", User{Username: Username("a"), Role: RoleAdmin}, InvalidUsernameError("username cannot be shorter than 2 characters")},
		{"invalid role", User{Username: Username("username"), Role: Role("invalid")}, ErrInvalidRole},
	}

	for _, c := range testCases {
		t.Run(c.desc, func(t *testing.T) {
			t.Parallel()
			err := c.user.IsValid()
			if !errors.Is(err, c.err) {
				t.Errorf("expected %q, got %q", c.err, err)
			}
		})
	}
}

func TestUsername_IsValid(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		desc     string
		username string
		err      error
	}{
		{"valid username", "valid", nil},
		{"username too short", "a", InvalidUsernameError("username cannot be shorter than 2 characters")},
		{"username too long", strings.Repeat("a", 256), InvalidUsernameError("username cannot be longer than 255 characters")},
		{"username contains disallowed characters", "####", InvalidUsernameError(`username can only contain alphanumeric characters, "-" and "_"`)},
	}

	for _, c := range testCases {
		t.Run(c.desc, func(t *testing.T) {
			t.Parallel()
			err := Username(c.username).IsValid()
			if !errors.Is(err, c.err) {
				t.Errorf("expected %q, got %q", c.err, err)
			}
		})
	}
}

func TestRole_IsValid(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		desc string
		role string
		err  error
	}{
		{"admin", "admin", nil},
		{"user", "user", nil},
		{"invalid role", "invalid", ErrInvalidRole},
	}

	for _, c := range testCases {
		t.Run(c.desc, func(t *testing.T) {
			t.Parallel()
			err := Role(c.role).IsValid()
			if !errors.Is(err, c.err) {
				t.Errorf("expected %q, got %q", c.err, err)
			}
		})
	}
}
