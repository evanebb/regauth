package repository

import "errors"

var (
	ErrNotFound          = errors.New("repository not found")
	ErrAlreadyExists     = errors.New("repository already exists, cannot create it again")
	ErrInvalidVisibility = errors.New("visibility is not valid, must be one of 'public', 'private'")
)

type InvalidNameError string

func (e InvalidNameError) Error() string {
	return "invalid repository name: " + string(e)
}
