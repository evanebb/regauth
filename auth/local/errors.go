package local

import "errors"

var (
	ErrNoCredentials = errors.New("no credentials found for user")
	ErrWeakPassword  = errors.New("password is too weak, must be at least 8 characters")
)
