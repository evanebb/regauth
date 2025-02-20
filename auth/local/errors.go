package local

import "errors"

var (
	ErrUserNotFound      = errors.New("auth user not found")
	ErrUserAlreadyExists = errors.New("auth user already exists, cannot create it again")
)
