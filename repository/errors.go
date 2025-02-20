package repository

import "errors"

var (
	ErrNotFound      = errors.New("repository not found")
	ErrAlreadyExists = errors.New("repository already exists, cannot create it again")
)
