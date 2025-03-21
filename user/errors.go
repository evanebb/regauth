package user

import "errors"

var (
	ErrNotFound           = errors.New("user not found")
	ErrAlreadyExists      = errors.New("user already exists, cannot create it again")
	ErrTeamNotFound       = errors.New("team not found")
	ErrTeamAlreadyExists  = errors.New("team already exists, cannot create it again")
	ErrTeamMemberNotFound = errors.New("team member not found")
)
