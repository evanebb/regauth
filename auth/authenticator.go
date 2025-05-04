package auth

import (
	"context"
	"errors"
	"github.com/evanebb/regauth/token"
	"github.com/evanebb/regauth/user"
	"net"
	"time"
)

type Authenticator interface {
	Authenticate(ctx context.Context, username, password string, sourceIP net.IP) (token.PersonalAccessToken, user.User, error)
}

func NewAuthenticator(t token.Store, u user.Store, tokenPrefix string) Authenticator {
	return authenticator{tokenStore: t, userStore: u, tokenPrefix: tokenPrefix}
}

type authenticator struct {
	tokenStore  token.Store
	userStore   user.Store
	tokenPrefix string
}

func (a authenticator) Authenticate(ctx context.Context, username, password string, sourceIP net.IP) (token.PersonalAccessToken, user.User, error) {
	var u user.User
	var p token.PersonalAccessToken

	u, err := a.userStore.GetByUsername(ctx, username)
	if err != nil {
		return p, u, errors.Join(ErrAuthenticationFailed, err)
	}

	p, err = a.tokenStore.GetByPlainTextToken(ctx, password)
	if err != nil {
		if errors.Is(err, token.ErrNotFound) {
			err = errors.Join(ErrAuthenticationFailed, err)
		}

		return p, u, err
	}

	if p.UserID != u.ID {
		return p, u, errors.Join(ErrAuthenticationFailed, errors.New("token does not belong to user"))
	}

	if p.ExpirationDate.Before(time.Now()) {
		return p, u, errors.Join(ErrAuthenticationFailed, errors.New("token has expired"))
	}

	// Log that the token was used
	logEntry := token.UsageLogEntry{
		TokenID:   p.ID,
		SourceIP:  sourceIP,
		Timestamp: time.Now(),
	}

	err = a.tokenStore.AddUsageLogEntry(ctx, logEntry)
	return p, u, err
}
