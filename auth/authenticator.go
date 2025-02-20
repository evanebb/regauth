package auth

import (
	"context"
	"errors"
	"github.com/evanebb/regauth/pat"
	"github.com/evanebb/regauth/user"
	"golang.org/x/crypto/bcrypt"
	"net"
	"strings"
	"time"
)

type Authenticator interface {
	Authenticate(ctx context.Context, username, password string, sourceIP net.IP) (pat.PersonalAccessToken, user.User, error)
}

func NewAuthenticator(p pat.Store, u user.Store) Authenticator {
	return authenticator{patStore: p, userStore: u}
}

type authenticator struct {
	patStore  pat.Store
	userStore user.Store
}

func (a authenticator) Authenticate(ctx context.Context, username, password string, sourceIP net.IP) (pat.PersonalAccessToken, user.User, error) {
	var u user.User
	var p pat.PersonalAccessToken

	u, err := a.userStore.GetByUsername(ctx, username)
	if err != nil {
		return p, u, errors.Join(ErrAuthenticationFailed, err)
	}

	if !strings.HasPrefix(password, "registry_pat_") {
		return p, u, errors.Join(ErrAuthenticationFailed, errors.New("invalid personal access token given"))
	}

	p, err = a.checkTokenAgainstUserTokens(ctx, u, password)
	if err == nil {
		logEntry := pat.UsageLogEntry{
			TokenID:   p.ID,
			SourceIP:  sourceIP,
			Timestamp: time.Now(),
		}

		err = a.patStore.AddUsageLogEntry(ctx, logEntry)
	}

	return p, u, err
}

func (a authenticator) checkTokenAgainstUserTokens(ctx context.Context, u user.User, token string) (pat.PersonalAccessToken, error) {
	var p pat.PersonalAccessToken

	userTokens, err := a.patStore.GetAllForUser(ctx, u.ID)
	if err != nil {
		return p, errors.Join(ErrAuthenticationFailed, err)
	}

	now := time.Now()
	for _, userToken := range userTokens {
		if userToken.ExpirationDate.Before(now) {
			continue
		}

		if err := bcrypt.CompareHashAndPassword(userToken.Hash, []byte(token)); err != nil {
			continue
		}

		return userToken, nil
	}

	return p, errors.Join(ErrAuthenticationFailed, errors.New("no matching personal access token found"))
}
