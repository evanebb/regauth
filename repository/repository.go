package repository

import (
	"errors"
	"github.com/google/uuid"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

type Repository struct {
	ID         uuid.UUID
	Namespace  string
	Name       string
	Visibility Visibility
	OwnerID    uuid.UUID
}

func (r Repository) IsValid() error {
	return r.Visibility.IsValid()
}

type Visibility string

func (v Visibility) IsValid() error {
	if v != Public && v != Private {
		return errors.New("visibility is not valid, must be one of 'public', 'private'")
	}

	return nil
}

func (v Visibility) String() string {
	return string(v)
}

func (v Visibility) HumanReadable() string {
	return cases.Title(language.Und).String(v.String())
}

const (
	Public  Visibility = "public"
	Private Visibility = "private"
)
