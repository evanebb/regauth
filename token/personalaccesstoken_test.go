package token

import (
	"errors"
	"strings"
	"testing"
)

func TestPersonalAccessToken_IsValid(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		desc  string
		token PersonalAccessToken
		err   error
	}{
		{"valid token", PersonalAccessToken{Description: Description("description"), Permission: PermissionReadOnly}, nil},
		{"invalid description", PersonalAccessToken{Description: Description("a"), Permission: PermissionReadOnly}, InvalidDescriptionError("description cannot be shorter than 2 characters")},
		{"invalid permission", PersonalAccessToken{Description: Description("description"), Permission: Permission("invalid")}, ErrInvalidPermission},
	}

	for _, c := range testCases {
		t.Run(c.desc, func(t *testing.T) {
			err := c.token.IsValid()
			if !errors.Is(err, c.err) {
				t.Errorf("expected %q, got %q", c.err, err)
			}
		})
	}
}

func TestDescription_IsValid(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		desc        string
		description string
		err         error
	}{
		{"valid description", "this is a valid description", nil},
		{"description too short", "a", InvalidDescriptionError("description cannot be shorter than 2 characters")},
		{"description too long", strings.Repeat("a", 256), InvalidDescriptionError("description cannot be longer than 255 characters")},
		{"description contains disallowed characters", "####", InvalidDescriptionError(`description can only contain alphanumeric characters, spaces, "-" and "_"`)},
	}

	for _, c := range testCases {
		t.Run(c.desc, func(t *testing.T) {
			err := Description(c.description).IsValid()
			if !errors.Is(err, c.err) {
				t.Errorf("expected %q, got %q", c.err, err)
			}
		})
	}
}

func TestPermission_IsValid(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		desc       string
		permission string
		err        error
	}{
		{"readOnly", "readOnly", nil},
		{"readWrite", "readWrite", nil},
		{"readWriteDelete", "readWriteDelete", nil},
		{"invalid permission", "invalid", ErrInvalidPermission},
	}

	for _, c := range testCases {
		t.Run(c.desc, func(t *testing.T) {
			err := Permission(c.permission).IsValid()
			if !errors.Is(err, c.err) {
				t.Errorf("expected %q, got %q", c.err, err)
			}
		})
	}
}
