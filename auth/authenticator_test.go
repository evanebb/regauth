package auth

import (
	"errors"
	"github.com/evanebb/regauth/store/memory"
	"github.com/evanebb/regauth/token"
	"github.com/evanebb/regauth/user"
	"github.com/google/uuid"
	"net"
	"strings"
	"testing"
	"time"
)

func TestAuthenticator_Authenticate(t *testing.T) {
	t.Parallel()

	sourceIP := net.ParseIP("192.168.10.1")

	t.Run("user does not exist", func(t *testing.T) {
		t.Parallel()
		tokenStore := memory.NewPersonalAccessTokenStore()
		userStore := memory.NewUserStore()
		a := NewAuthenticator(tokenStore, userStore, "registry_pat_")

		_, _, err := a.Authenticate(t.Context(), "foo", "", sourceIP)
		if !errors.Is(err, ErrAuthenticationFailed) || !errors.Is(err, user.ErrNotFound) {
			t.Fatalf("expected %q, %q, got %q", ErrAuthenticationFailed, user.ErrNotFound, err)
		}
	})

	t.Run("personal access token does not exist", func(t *testing.T) {
		t.Parallel()
		tokenStore := memory.NewPersonalAccessTokenStore()
		userStore := memory.NewUserStore()
		a := NewAuthenticator(tokenStore, userStore, "registry_pat_")

		u := user.User{
			ID:       uuid.New(),
			Username: "user",
			Role:     user.RoleAdmin,
		}
		if err := userStore.Create(t.Context(), u); err != nil {
			t.Fatalf("failed to create user: %q", err)
		}

		_, _, err := a.Authenticate(t.Context(), "user", "registry_pat_foobarbaz", sourceIP)
		if !errors.Is(err, ErrAuthenticationFailed) || !errors.Is(err, token.ErrNotFound) {
			t.Fatalf("expected %q, %q, got %q", ErrAuthenticationFailed, token.ErrNotFound, err)
		}
	})

	t.Run("personal access token belongs to different user", func(t *testing.T) {
		t.Parallel()
		tokenStore := memory.NewPersonalAccessTokenStore()
		userStore := memory.NewUserStore()
		a := NewAuthenticator(tokenStore, userStore, "registry_pat_")

		u := user.User{
			ID:       uuid.New(),
			Username: "user",
			Role:     user.RoleAdmin,
		}
		if err := userStore.Create(t.Context(), u); err != nil {
			t.Fatalf("failed to create user: %q", err)
		}

		tok := token.PersonalAccessToken{
			ID:             uuid.New(),
			Description:    "token",
			Permission:     token.PermissionReadOnly,
			ExpirationDate: time.Now(),
			UserID:         uuid.New(),
		}
		if err := tokenStore.Create(t.Context(), tok, "registry_pat_foobarbaz"); err != nil {
			t.Fatalf("failed to create personal access token: %q", err)
		}

		_, _, err := a.Authenticate(t.Context(), "user", "registry_pat_foobarbaz", sourceIP)
		if !errors.Is(err, ErrAuthenticationFailed) || !strings.Contains(err.Error(), "token does not belong to user") {
			t.Fatalf("expected %q, got %q", ErrAuthenticationFailed, err)
		}
	})

	t.Run("personal access token has expired", func(t *testing.T) {
		t.Parallel()
		tokenStore := memory.NewPersonalAccessTokenStore()
		userStore := memory.NewUserStore()
		a := NewAuthenticator(tokenStore, userStore, "registry_pat_")

		u := user.User{
			ID:       uuid.New(),
			Username: "user",
			Role:     user.RoleAdmin,
		}
		if err := userStore.Create(t.Context(), u); err != nil {
			t.Fatalf("failed to create user: %q", err)
		}

		tok := token.PersonalAccessToken{
			ID:             uuid.New(),
			Description:    "token",
			Permission:     token.PermissionReadOnly,
			ExpirationDate: time.Now().Add(-time.Minute),
			UserID:         u.ID,
		}
		if err := tokenStore.Create(t.Context(), tok, "registry_pat_foobarbaz"); err != nil {
			t.Fatalf("failed to create personal access token: %q", err)
		}

		_, _, err := a.Authenticate(t.Context(), "user", "registry_pat_foobarbaz", sourceIP)
		if !errors.Is(err, ErrAuthenticationFailed) || !strings.Contains(err.Error(), "token has expired") {
			t.Fatalf("expected %q, got %q", ErrAuthenticationFailed, err)
		}
	})

	t.Run("successful authentication", func(t *testing.T) {
		t.Parallel()
		tokenStore := memory.NewPersonalAccessTokenStore()
		userStore := memory.NewUserStore()
		a := NewAuthenticator(tokenStore, userStore, "registry_pat_")

		u := user.User{
			ID:       uuid.New(),
			Username: "user",
			Role:     user.RoleAdmin,
		}
		if err := userStore.Create(t.Context(), u); err != nil {
			t.Fatalf("failed to create user: %q", err)
		}

		tok := token.PersonalAccessToken{
			ID:             uuid.New(),
			Description:    "token",
			Permission:     token.PermissionReadOnly,
			ExpirationDate: time.Now().Add(time.Hour),
			UserID:         u.ID,
		}
		if err := tokenStore.Create(t.Context(), tok, "registry_pat_foobarbaz"); err != nil {
			t.Fatalf("failed to create personal access token: %q", err)
		}

		actualToken, actualUser, err := a.Authenticate(t.Context(), "user", "registry_pat_foobarbaz", sourceIP)
		if err != nil {
			t.Fatalf("expected err to be nil, got %q", err)
		}

		if actualToken != tok || actualUser != u {
			t.Fatalf("expected %+v, %+v, got %+v, %+v", tok, u, actualToken, actualUser)
		}

		// verify that an entry was added to the usage log
		usageLog, err := tokenStore.GetUsageLog(t.Context(), tok.ID)
		if err != nil {
			t.Fatalf("expected err to be nil, got %q", err)
		}

		if len(usageLog) != 1 {
			t.Fatalf("expected one token usage log entry, got %d entries", len(usageLog))
		}

		entry := usageLog[0]
		if entry.TokenID != tok.ID || !entry.SourceIP.Equal(sourceIP) {
			t.Fatalf("expected log entry with token ID %q, source IP %q, got %q, %q",
				tok.ID, sourceIP.String(), entry.TokenID, entry.SourceIP.String())
		}
	})
}
