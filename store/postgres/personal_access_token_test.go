package postgres

import (
	"errors"
	"github.com/evanebb/regauth/token"
	"github.com/google/uuid"
	"net"
	"testing"
	"time"
)

func TestPersonalAccessTokenStore_GetAllByUser(t *testing.T) {
	db := getDatabaseConnection(t)
	s := NewPersonalAccessTokenStore(db)

	userID, _ := uuid.Parse("0195cd11-2863-71d4-a3c4-032bc264cf81")

	tokens, err := s.GetAllByUser(t.Context(), userID)
	if err != nil {
		t.Errorf("expected err to be nil, got %q", err)
	}

	// the seeds create three personal access tokens for this user, so we expect there to be three
	if len(tokens) != 3 {
		t.Errorf("expected three tokens, got %d", len(tokens))
	}
}

func TestPersonalAccessTokenStore_GetByID(t *testing.T) {
	db := getDatabaseConnection(t)
	s := NewPersonalAccessTokenStore(db)

	t.Run("existing token", func(t *testing.T) {
		userID, _ := uuid.Parse("0195cd11-2863-71d4-a3c4-032bc264cf81")
		tokenID, _ := uuid.Parse("0195cd16-2142-78e5-8425-a8db7acbc8f8")
		expirationDate, _ := time.Parse(time.RFC3339, "2045-03-25T12:16:33.110405Z")

		expectedToken := token.PersonalAccessToken{
			ID:             tokenID,
			Description:    "Read-only token",
			Permission:     token.PermissionReadOnly,
			ExpirationDate: expirationDate,
			UserID:         userID,
		}

		tok, err := s.GetByID(t.Context(), tokenID)
		if err != nil {
			t.Errorf("expected err to be nil, got %q", err)
		}

		if tok != expectedToken {
			t.Errorf("expected %+v, got %+v", expectedToken, tok)
		}
	})

	t.Run("token does not exist", func(t *testing.T) {
		if _, err := s.GetByID(t.Context(), uuid.Nil); !errors.Is(err, token.ErrNotFound) {
			t.Errorf("expected %q, got %q", token.ErrNotFound, err)
		}
	})
}

func TestPersonalAccessTokenStore_GetByPlainTextToken(t *testing.T) {
	db := getDatabaseConnection(t)
	s := NewPersonalAccessTokenStore(db)

	t.Run("existing token", func(t *testing.T) {
		userID, _ := uuid.Parse("0195cd11-2863-71d4-a3c4-032bc264cf81")
		tokenID, _ := uuid.Parse("0195cd16-2142-78e5-8425-a8db7acbc8f8")
		expirationDate, _ := time.Parse(time.RFC3339, "2045-03-25T12:16:33.110405Z")

		expectedToken := token.PersonalAccessToken{
			ID:             tokenID,
			Description:    "Read-only token",
			Permission:     token.PermissionReadOnly,
			ExpirationDate: expirationDate,
			UserID:         userID,
		}

		plainTextToken := "registry_pat_SVV_otfQNmSjo7viDiCrC0AKe6Qa_iFhxXJBZE1vMOByC9nbUtBPsz3r"
		tok, err := s.GetByPlainTextToken(t.Context(), plainTextToken)
		if err != nil {
			t.Errorf("expected err to be nil, got %q", err)
		}

		if tok != expectedToken {
			t.Errorf("expected %+v, got %+v", expectedToken, tok)
		}
	})

	t.Run("invalid token, too short", func(t *testing.T) {
		if _, err := s.GetByPlainTextToken(t.Context(), "foo"); !errors.Is(err, ErrTokenTooShort) {
			t.Errorf("expected %q, got %q", ErrTokenTooShort, err)
		}
	})

	t.Run("token does not exist", func(t *testing.T) {
		if _, err := s.GetByPlainTextToken(t.Context(), "registry_pat_invalid"); !errors.Is(err, token.ErrNotFound) {
			t.Errorf("expected %q, got %q", token.ErrNotFound, err)
		}
	})
}

func TestPersonalAccessTokenStore_Create(t *testing.T) {
	db := getDatabaseConnection(t)
	s := NewPersonalAccessTokenStore(db)

	userID, _ := uuid.Parse("0195cd11-2863-71d4-a3c4-032bc264cf81")
	expirationDate, _ := time.Parse(time.RFC3339, "2045-03-25T12:16:33.110405Z")

	plainTextToken := "registry_pat_foobarbaz"
	tok := token.PersonalAccessToken{
		ID:             uuid.New(),
		Description:    "Read-only token",
		Permission:     token.PermissionReadOnly,
		ExpirationDate: expirationDate,
		UserID:         userID,
	}

	t.Run("new token", func(t *testing.T) {
		if err := s.Create(t.Context(), tok, plainTextToken); err != nil {
			t.Errorf("expected nil, got %q", err)
		}
	})

	t.Run("token already exists", func(t *testing.T) {
		if err := s.Create(t.Context(), tok, plainTextToken); !errors.Is(err, token.ErrAlreadyExists) {
			t.Errorf("expected %q, got %q", token.ErrAlreadyExists, err)
		}
	})
}

func TestPersonalAccessTokenStore_DeleteByID(t *testing.T) {
	db := getDatabaseConnection(t)
	s := NewPersonalAccessTokenStore(db)

	tokenID, _ := uuid.Parse("0195cd16-2142-78e5-8425-a8db7acbc8f8")

	if _, err := s.GetByID(t.Context(), tokenID); err != nil {
		t.Errorf("expected token to exist, got %q", err)
	}

	if err := s.DeleteByID(t.Context(), tokenID); err != nil {
		t.Errorf("expected nil, got %q", err)
	}

	if _, err := s.GetByID(t.Context(), tokenID); !errors.Is(err, token.ErrNotFound) {
		t.Errorf("expected token to be deleted, got %q", err)
	}
}

func TestPersonalAccessTokenStore_GetUsageLog(t *testing.T) {
	db := getDatabaseConnection(t)
	s := NewPersonalAccessTokenStore(db)

	tokenID, _ := uuid.Parse("0195cd16-2142-78e5-8425-a8db7acbc8f8")

	entries, err := s.GetUsageLog(t.Context(), tokenID)
	if err != nil {
		t.Errorf("expected nil, got %q", err)
	}

	if len(entries) != 3 {
		t.Errorf("expected three log entries, got %d", len(entries))
	}
}

func TestPersonalAccessTokenStore_AddUsageLogEntry(t *testing.T) {
	db := getDatabaseConnection(t)
	s := NewPersonalAccessTokenStore(db)

	tokenID, _ := uuid.Parse("0195cd16-2142-78e5-8425-a8db7acbc8f8")
	entry := token.UsageLogEntry{
		TokenID:   tokenID,
		SourceIP:  net.ParseIP("192.168.1.10"),
		Timestamp: time.Now(),
	}

	if err := s.AddUsageLogEntry(t.Context(), entry); err != nil {
		t.Errorf("expected nil, got %q", err)
	}

	entries, err := s.GetUsageLog(t.Context(), tokenID)
	if err != nil {
		t.Errorf("expected nil, got %q", err)
	}

	if len(entries) != 4 {
		t.Errorf("expected four log entries, got %d", len(entries))
	}
}
