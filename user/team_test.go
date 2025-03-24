package user

import (
	"errors"
	"github.com/google/uuid"
	"strings"
	"testing"
)

func TestTeam_IsValid(t *testing.T) {
	testCases := []struct {
		desc string
		team Team
		err  error
	}{
		{"valid team", Team{ID: uuid.New(), Name: TeamName("team")}, nil},
		{"invalid name", Team{ID: uuid.New(), Name: TeamName("a")}, InvalidTeamNameError("team name cannot be shorter than 2 characters")},
	}

	for _, c := range testCases {
		t.Run(c.desc, func(t *testing.T) {
			err := c.team.IsValid()
			if !errors.Is(err, c.err) {
				t.Errorf("expected %q, got %q", c.err, err)
			}
		})
	}
}

func TestTeamName_IsValid(t *testing.T) {
	testCases := []struct {
		desc string
		name string
		err  error
	}{
		{"valid team name", "valid", nil},
		{"team name too short", "a", InvalidTeamNameError("team name cannot be shorter than 2 characters")},
		{"team name too long", strings.Repeat("a", 256), InvalidTeamNameError("team name cannot be longer than 255 characters")},
		{"team name contains disallowed characters", "####", InvalidTeamNameError(`team name can only contain alphanumeric characters, "-" and "_"`)},
	}

	for _, c := range testCases {
		t.Run(c.desc, func(t *testing.T) {
			err := TeamName(c.name).IsValid()
			if !errors.Is(err, c.err) {
				t.Errorf("expected %q, got %q", c.err, err)
			}
		})
	}
}

func TestTeamMember_IsValid(t *testing.T) {
	testCases := []struct {
		desc   string
		member TeamMember
		err    error
	}{
		{"valid user", TeamMember{UserID: uuid.New(), TeamID: uuid.New(), Username: Username("username"), Role: TeamMemberRoleAdmin}, nil},
		{"invalid username", TeamMember{UserID: uuid.New(), TeamID: uuid.New(), Username: Username("a"), Role: TeamMemberRoleAdmin}, InvalidUsernameError("username cannot be shorter than 2 characters")},
		{"invalid role", TeamMember{UserID: uuid.New(), TeamID: uuid.New(), Username: Username("username"), Role: TeamMemberRole("invalid")}, ErrInvalidTeamMemberRole},
	}

	for _, c := range testCases {
		t.Run(c.desc, func(t *testing.T) {
			err := c.member.IsValid()
			if !errors.Is(err, c.err) {
				t.Errorf("expected %q, got %q", c.err, err)
			}
		})
	}
}

func TestTeamMemberRole_IsValid(t *testing.T) {
	testCases := []struct {
		desc string
		role string
		err  error
	}{
		{"admin", "admin", nil},
		{"user", "user", nil},
		{"invalid role", "invalid", ErrInvalidTeamMemberRole},
	}

	for _, c := range testCases {
		t.Run(c.desc, func(t *testing.T) {
			err := TeamMemberRole(c.role).IsValid()
			if !errors.Is(err, c.err) {
				t.Errorf("expected %q, got %q", c.err, err)
			}
		})
	}
}
