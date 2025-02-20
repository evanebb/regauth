package pat

import "errors"

var (
	ErrNotFound      = errors.New("personal access token not found")
	ErrAlreadyExists = errors.New("personal access token already exists, cannot create it again")
)
