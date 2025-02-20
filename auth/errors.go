package auth

import "errors"

var ErrAuthenticationFailed = errors.New("authentication failed")
var ErrAccessNotGranted = errors.New("access to the requested resource was not granted")
