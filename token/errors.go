package token

import "errors"

var (
	ErrNotFound          = errors.New("personal access token not found")
	ErrAlreadyExists     = errors.New("personal access token already exists, cannot create it again")
	ErrInvalidPermission = errors.New("permission is not valid, must be one of 'readOnly', 'readWrite', 'readWriteDelete'")
)

type InvalidDescriptionError string

func (e InvalidDescriptionError) Error() string {
	return "invalid personal access token description: " + string(e)
}
