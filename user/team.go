package user

import (
	"errors"
	"github.com/google/uuid"
)

type Team struct {
	ID   uuid.UUID `json:"id"`
	Name string    `json:"name"`
}

type TeamMember struct {
	UserID   uuid.UUID      `json:"userId"`
	TeamID   uuid.UUID      `json:"-"`
	Username Username       `json:"username"`
	Role     TeamMemberRole `json:"role"`
}

func (m TeamMember) IsValid() error {
	if err := m.Username.IsValid(); err != nil {
		return err
	}

	if err := m.Role.IsValid(); err != nil {
		return err
	}

	return nil
}

type TeamMemberRole string

const (
	TeamMemberRoleAdmin TeamMemberRole = "admin"
	TeamMemberRoleUser  TeamMemberRole = "user"
)

func (r TeamMemberRole) IsValid() error {
	if r != TeamMemberRoleAdmin && r != TeamMemberRoleUser {
		return errors.New("team member role is not valid, must be one of 'admin', 'user'")
	}

	return nil
}
