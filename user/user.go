package user

import (
	"errors"
	"github.com/google/uuid"
	"regexp"
)

type User struct {
	ID       uuid.UUID `json:"id"`
	Username Username  `json:"username"`
	Role     Role      `json:"role"`
}

func (u User) IsValid() error {
	if err := u.Username.IsValid(); err != nil {
		return err
	}

	if err := u.Role.IsValid(); err != nil {
		return err
	}

	return nil
}

type Username string

var validUsername = regexp.MustCompile(`^[a-zA-Z0-9-_]+$`)

func (u Username) IsValid() error {
	l := len([]rune(u))
	if l < 2 {
		return errors.New("username cannot be shorter than 2 characters")
	}

	if l > 255 {
		return errors.New("username cannot be longer than 255 characters")
	}

	if !validUsername.MatchString(string(u)) {
		return errors.New(`username can only contain alphanumeric characters, "-" and "_"`)
	}

	return nil
}

func (u Username) String() string {
	return string(u)
}

type Role string

const (
	RoleAdmin Role = "admin"
	RoleUser  Role = "user"
)

func (r Role) IsValid() error {
	if r != RoleAdmin && r != RoleUser {
		return errors.New("role is not valid, must be one of 'admin', 'user'")
	}

	return nil
}

func (r Role) String() string {
	return string(r)
}
