package postgres

import (
	"context"
	"errors"
	"github.com/evanebb/regauth/auth/local"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type AuthUserStore struct {
	TransactionStore
}

func NewAuthUserStore(db *pgxpool.Pool) local.AuthUserStore {
	return AuthUserStore{TransactionStore{db: db}}
}

func (s AuthUserStore) GetByID(ctx context.Context, id uuid.UUID) (local.AuthUser, error) {
	var u local.AuthUser

	query := "SELECT uuid, username, password_hash FROM local_auth_users WHERE uuid = $1"
	err := s.QuerierFromContext(ctx).QueryRow(ctx, query, id).Scan(&u.ID, &u.Username, &u.PasswordHash)
	if errors.Is(err, pgx.ErrNoRows) {
		return u, local.ErrUserNotFound
	}

	return u, err
}

func (s AuthUserStore) GetByUsername(ctx context.Context, username string) (local.AuthUser, error) {
	var u local.AuthUser

	query := "SELECT uuid, username, password_hash FROM local_auth_users WHERE username = $1"
	err := s.QuerierFromContext(ctx).QueryRow(ctx, query, username).Scan(&u.ID, &u.Username, &u.PasswordHash)
	if errors.Is(err, pgx.ErrNoRows) {
		return u, local.ErrUserNotFound
	}

	return u, err
}

func (s AuthUserStore) Create(ctx context.Context, u local.AuthUser) error {
	_, err := s.GetByID(ctx, u.ID)
	if err == nil {
		return local.ErrUserAlreadyExists
	}
	if !errors.Is(err, local.ErrUserNotFound) {
		return err
	}

	query := "INSERT INTO local_auth_users (uuid, username, password_hash) VALUES ($1, $2, $3)"
	_, err = s.QuerierFromContext(ctx).Exec(ctx, query, u.ID, u.Username, u.PasswordHash)
	return err
}

func (s AuthUserStore) Update(ctx context.Context, u local.AuthUser) error {
	_, err := s.GetByID(ctx, u.ID)
	if err != nil {
		return err
	}

	query := "UPDATE local_auth_users SET username = $1, password_hash = $2 WHERE uuid = $3"
	_, err = s.QuerierFromContext(ctx).Exec(ctx, query, u.Username, u.PasswordHash, u.ID)
	return err
}

func (s AuthUserStore) DeleteByID(ctx context.Context, id uuid.UUID) error {
	query := "DELETE FROM local_auth_users WHERE uuid = $1"
	_, err := s.QuerierFromContext(ctx).Exec(ctx, query, id)
	return err
}
