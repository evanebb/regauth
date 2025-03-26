package user

import "errors"

var (
	ErrNotFound                = errors.New("user not found")
	ErrAlreadyExists           = errors.New("user already exists, cannot create it again")
	ErrInvalidRole             = errors.New("role is not valid, must be one of 'admin', 'user'")
	ErrTeamNotFound            = errors.New("team not found")
	ErrTeamAlreadyExists       = errors.New("team already exists, cannot create it again")
	ErrTeamMemberNotFound      = errors.New("team member not found")
	ErrTeamMemberAlreadyExists = errors.New("team member already exists in team")
	ErrInvalidTeamMemberRole   = errors.New("team member role is not valid, must be one of 'admin', 'user'")
)

type InvalidUsernameError string

func (e InvalidUsernameError) Error() string {
	return "invalid username: " + string(e)
}

type InvalidTeamNameError string

func (e InvalidTeamNameError) Error() string {
	return "invalid team name: " + string(e)
}
