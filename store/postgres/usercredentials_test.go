package postgres

import (
	"bytes"
	"errors"
	"github.com/evanebb/regauth/auth/local"
	"github.com/google/uuid"
	"testing"
)

func TestUserCredentialsStore_GetByUserID(t *testing.T) {
	db := getDatabaseConnection(t)
	s := NewUserCredentialsStore(db)

	t.Run("existing user credentials", func(t *testing.T) {
		userID, _ := uuid.Parse("0195cd11-2863-71d4-a3c4-032bc264cf81")
		expectedHash := []byte("$2y$12$sSMlPGBCt2RZnX5Od405T./kEwKZYtJoIhijrL1XXlwvr/BtPDtgS")

		credentials, err := s.GetByUserID(t.Context(), userID)
		if err != nil {
			t.Errorf("expected err to be nil, got %q", err)
		}

		if credentials.UserID != userID || !bytes.Equal(credentials.PasswordHash, expectedHash) {
			t.Errorf("unexpected credentials, got %+v", credentials)
		}
	})

	t.Run("user has no credentials", func(t *testing.T) {
		if _, err := s.GetByUserID(t.Context(), uuid.Nil); !errors.Is(err, local.ErrNoCredentials) {
			t.Errorf("expected %q, got %q", local.ErrNoCredentials, err)
		}
	})
}

func TestUserCredentialsStore_Save(t *testing.T) {
	db := getDatabaseConnection(t)
	s := NewUserCredentialsStore(db)

	userID, _ := uuid.Parse("0195cd11-2863-71d4-a3c4-032bc264cf81")
	hash := []byte("$2a$12$ORC.TlUFDczMOeSbB59W8.ViFE/b47o/Yur6qan4hnwch.Wqa7Hty")

	credentials := local.UserCredentials{
		UserID:       userID,
		PasswordHash: hash,
	}

	if err := s.Save(t.Context(), credentials); err != nil {
		t.Errorf("expected nil, got %q", err)
	}
}
