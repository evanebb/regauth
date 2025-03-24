package user

import (
	"github.com/google/uuid"
	"regexp"
)

type Team struct {
	ID   uuid.UUID `json:"id"`
	Name TeamName  `json:"name"`
}

func (t Team) IsValid() error {
	if err := t.Name.IsValid(); err != nil {
		return err
	}

	return nil
}

type TeamName string

var validTeamName = regexp.MustCompile(`^[a-zA-Z0-9-_]+$`)

func (n TeamName) IsValid() error {
	l := len([]rune(n))
	if l < 2 {
		return InvalidTeamNameError("team name cannot be shorter than 2 characters")
	}

	if l > 255 {
		return InvalidTeamNameError("team name cannot be longer than 255 characters")
	}

	if !validTeamName.MatchString(string(n)) {
		return InvalidTeamNameError(`team name can only contain alphanumeric characters, "-" and "_"`)
	}

	return nil
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
		return ErrInvalidTeamMemberRole
	}

	return nil
}
