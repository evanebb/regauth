package repository

import (
	"errors"
	"fmt"
	"strings"
	"testing"
)

func TestRepository_IsValid(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		desc string
		repo Repository
		err  error
	}{
		{"valid repository", Repository{Namespace: "namespace", Name: Name("name"), Visibility: VisibilityPrivate}, nil},
		{"invalid name", Repository{Namespace: "namespace", Name: Name("a"), Visibility: VisibilityPrivate}, InvalidNameError("name cannot be shorter than 2 characters")},
		{"invalid visibility", Repository{Namespace: "namespace", Name: Name("name"), Visibility: Visibility("invalid")}, ErrInvalidVisibility},
	}

	for _, c := range testCases {
		t.Run(c.desc, func(t *testing.T) {
			t.Parallel()
			err := c.repo.IsValid()
			if !errors.Is(err, c.err) {
				t.Errorf("expected %q, got %q", c.err, err)
			}
		})
	}
}

type nameTestCase struct {
	desc string
	name string
	err  error
}

func TestName_IsValid(t *testing.T) {
	t.Parallel()

	testCases := []nameTestCase{
		{"valid name", "valid", nil},
		{"name too short", "a", InvalidNameError("name cannot be shorter than 2 characters")},
		{"name too long", strings.Repeat("a", 256), InvalidNameError("name cannot be longer than 255 characters")},
	}

	for _, char := range []string{"-", "_", "."} {
		testCases = append(testCases,
			nameTestCase{
				desc: fmt.Sprintf(`disallowed start character "%s"`, char),
				name: char + "invalid",
				err:  InvalidNameError(`name cannot start or end with "-", "_" or "."`),
			},
			nameTestCase{
				desc: fmt.Sprintf(`disallowed end character "%s"`, char),
				name: "invalid" + char,
				err:  InvalidNameError(`name cannot start or end with "-", "_" or "."`),
			},
			nameTestCase{
				desc: fmt.Sprintf(`disallowed repeating character "%s"`, char),
				name: "foo" + strings.Repeat(char, 2) + "bar",
				err:  InvalidNameError(`name cannot contain repeating "-", "_" and "." characters`),
			},
		)
	}

	for _, c := range testCases {
		t.Run(c.desc, func(t *testing.T) {
			t.Parallel()
			err := Name(c.name).IsValid()
			if !errors.Is(err, c.err) {
				t.Errorf("expected %q, got %q", c.err, err)
			}
		})
	}
}

func TestVisibility_IsValid(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		desc       string
		visibility string
		err        error
	}{
		{"private", "private", nil},
		{"public", "public", nil},
		{"invalid visibility", "invalid", ErrInvalidVisibility},
	}

	for _, c := range testCases {
		t.Run(c.desc, func(t *testing.T) {
			t.Parallel()
			err := Visibility(c.visibility).IsValid()
			if !errors.Is(err, c.err) {
				t.Errorf("expected %q, got %q", c.err, err)
			}
		})
	}
}
