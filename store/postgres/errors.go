package postgres

import "errors"

var (
	ErrTokenTooShort = errors.New("invalid token given, too short")
)
