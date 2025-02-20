package user

import (
	"github.com/google/uuid"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

type User struct {
	ID        uuid.UUID
	Username  string
	FirstName string
	LastName  string
	Role      Role
}

type Role string

const (
	RoleAdmin Role = "admin"
	RoleUser  Role = "user"
)

func (r Role) String() string {
	return string(r)
}

func (r Role) HumanReadable() string {
	return cases.Title(language.Und).String(r.String())
}
