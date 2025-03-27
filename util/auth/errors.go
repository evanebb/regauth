package auth

import "errors"

var (
	ErrHashDoesNotMatch = errors.New("token and hash do not match")
)
