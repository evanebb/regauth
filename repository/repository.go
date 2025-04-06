package repository

import (
	"github.com/google/uuid"
	"regexp"
	"strings"
	"time"
)

type Repository struct {
	ID         uuid.UUID
	Namespace  string
	Name       Name
	Visibility Visibility
	CreatedAt  time.Time
}

func (r Repository) IsValid() error {
	if err := r.Name.IsValid(); err != nil {
		return err
	}

	if err := r.Visibility.IsValid(); err != nil {
		return err
	}

	return nil
}

type Name string

var validName = regexp.MustCompile("^[a-zA-Z0-9-_.]+$")
var nameNoRepeatingSpecialChars = regexp.MustCompile("[-_.]{2,}")

func (n Name) IsValid() error {
	l := len([]rune(n))
	if l < 2 {
		return InvalidNameError("name cannot be shorter than 2 characters")
	}

	if l > 255 {
		return InvalidNameError("name cannot be longer than 255 characters")
	}

	str := string(n)
	disallowedStartEndChars := []string{"-", "_", "."}
	for _, disallowed := range disallowedStartEndChars {
		if strings.HasPrefix(str, disallowed) || strings.HasSuffix(str, disallowed) {
			return InvalidNameError(`name cannot start or end with "-", "_" or "."`)
		}
	}

	if !validName.MatchString(str) {
		return InvalidNameError(`name can only contain alphanumeric characters, "-", "_" and "." (non-repeating)`)
	}

	if nameNoRepeatingSpecialChars.MatchString(str) {
		return InvalidNameError(`name cannot contain repeating "-", "_" and "." characters`)
	}

	return nil
}

type Visibility string

const (
	VisibilityPublic  Visibility = "public"
	VisibilityPrivate Visibility = "private"
)

func (v Visibility) IsValid() error {
	if v != VisibilityPublic && v != VisibilityPrivate {
		return ErrInvalidVisibility
	}

	return nil
}
