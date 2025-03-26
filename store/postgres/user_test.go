package postgres

import (
	"errors"
	"github.com/evanebb/regauth/user"
	"github.com/google/uuid"
	"testing"
)

func TestUserStore_GetAll(t *testing.T) {
	db := getDatabaseConnection(t)
	s := NewUserStore(db)

	users, err := s.GetAll(t.Context())
	if err != nil {
		t.Errorf("expected err to be nil, got %q", err)
	}

	// the seeds create two users, so we expect there to be two
	if len(users) != 2 {
		t.Errorf("expected two users, got %d", len(users))
	}
}

func TestUserStore_GetByID(t *testing.T) {
	db := getDatabaseConnection(t)
	s := NewUserStore(db)

	t.Run("existing user", func(t *testing.T) {
		userID, _ := uuid.Parse("0195cd11-2863-71d4-a3c4-032bc264cf81")
		expectedUser := user.User{ID: userID, Username: "adminuser", Role: user.RoleAdmin}

		u, err := s.GetByID(t.Context(), userID)
		if err != nil {
			t.Errorf("expected err to be nil, got %q", err)
		}

		if u != expectedUser {
			t.Errorf("expected %+v, got %+v", expectedUser, u)
		}
	})

	t.Run("user does not exist", func(t *testing.T) {
		if _, err := s.GetByID(t.Context(), uuid.Nil); !errors.Is(err, user.ErrNotFound) {
			t.Errorf("expected %q, got %q", user.ErrNotFound, err)
		}
	})
}

func TestUserStore_GetByUsername(t *testing.T) {
	db := getDatabaseConnection(t)
	s := NewUserStore(db)

	t.Run("existing user", func(t *testing.T) {
		userID, _ := uuid.Parse("0195cd11-2863-71d4-a3c4-032bc264cf81")
		username := "adminuser"
		expectedUser := user.User{ID: userID, Username: user.Username(username), Role: user.RoleAdmin}

		u, err := s.GetByUsername(t.Context(), username)
		if err != nil {
			t.Errorf("expected err to be nil, got %q", err)
		}

		if u != expectedUser {
			t.Errorf("expected %+v, got %+v", expectedUser, u)
		}
	})

	t.Run("user does not exist", func(t *testing.T) {
		if _, err := s.GetByUsername(t.Context(), "doesnotexist"); !errors.Is(err, user.ErrNotFound) {
			t.Errorf("expected %q, got %q", user.ErrNotFound, err)
		}
	})
}

func TestUserStore_Create(t *testing.T) {
	db := getDatabaseConnection(t)
	s := NewUserStore(db)

	u := user.User{ID: uuid.New(), Username: "username", Role: user.RoleAdmin}

	t.Run("new user", func(t *testing.T) {
		if err := s.Create(t.Context(), u); err != nil {
			t.Errorf("expected nil, got %q", err)
		}
	})

	t.Run("user already exists", func(t *testing.T) {
		if err := s.Create(t.Context(), u); !errors.Is(err, user.ErrAlreadyExists) {
			t.Errorf("expected %q, got %q", user.ErrAlreadyExists, err)
		}
	})
}

func TestUserStore_DeleteByID(t *testing.T) {
	db := getDatabaseConnection(t)
	s := NewUserStore(db)

	userID, _ := uuid.Parse("0195cd11-2863-71d4-a3c4-032bc264cf81")

	if _, err := s.GetByID(t.Context(), userID); err != nil {
		t.Errorf("expected user to exist, got %q", err)
	}

	if err := s.DeleteByID(t.Context(), userID); err != nil {
		t.Errorf("expected nil, got %q", err)
	}

	if _, err := s.GetByID(t.Context(), userID); !errors.Is(err, user.ErrNotFound) {
		t.Errorf("expected user to be deleted, got %q", err)
	}
}
