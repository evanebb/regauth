package repository

import (
	"errors"
	"github.com/google/uuid"
	"regexp"
	"strings"
)

type Repository struct {
	ID         uuid.UUID  `json:"id"`
	Namespace  string     `json:"namespace"`
	Name       Name       `json:"name"`
	Visibility Visibility `json:"visibility"`
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
		return errors.New("name cannot be shorter than 2 characters")
	}

	if l > 255 {
		return errors.New("name cannot be longer than 255 characters")
	}

	str := n.String()
	disallowedStartEndChars := []string{"-", "_", "."}
	for _, disallowed := range disallowedStartEndChars {
		if strings.HasPrefix(str, disallowed) || strings.HasSuffix(str, disallowed) {
			return errors.New(`name cannot start or end with "-", "_" or "."`)
		}
	}

	if !validName.MatchString(str) {
		return errors.New(`name can only contain alphanumeric characters, "-", "_" and "." (non-repeating)`)
	}

	if nameNoRepeatingSpecialChars.MatchString(str) {
		return errors.New(`name cannot containing repeating "-", "_" and "." characters`)
	}

	return nil
}

func (n Name) String() string {
	return string(n)
}

type Visibility string

const (
	VisibilityPublic  Visibility = "public"
	VisibilityPrivate Visibility = "private"
)

func (v Visibility) IsValid() error {
	if v != VisibilityPublic && v != VisibilityPrivate {
		return errors.New("visibility is not valid, must be one of 'public', 'private'")
	}

	return nil
}

func (v Visibility) String() string {
	return string(v)
}
