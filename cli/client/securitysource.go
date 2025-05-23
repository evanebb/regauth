package client

import (
	"context"
	"github.com/evanebb/regauth/oas"
	"github.com/ogen-go/ogen/ogenerrors"
)

type SecuritySource struct {
	Token    string
	Username string
	Password string
}

func (s SecuritySource) PersonalAccessToken(ctx context.Context, operationName oas.OperationName) (oas.PersonalAccessToken, error) {
	if s.Token != "" {
		return oas.PersonalAccessToken{Token: s.Token}, nil
	}

	return oas.PersonalAccessToken{}, ogenerrors.ErrSkipClientSecurity
}

func (s SecuritySource) UsernamePassword(ctx context.Context, operationName oas.OperationName) (oas.UsernamePassword, error) {
	if s.Username != "" && s.Password != "" {
		return oas.UsernamePassword{Username: s.Username, Password: s.Password}, nil
	}

	return oas.UsernamePassword{}, ogenerrors.ErrSkipClientSecurity
}
