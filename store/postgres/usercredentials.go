package postgres

import (
	"context"
	"errors"
	"github.com/evanebb/regauth/auth/local"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type UserCredentialsStore struct {
	TransactionStore
}

func NewUserCredentialsStore(db *pgxpool.Pool) UserCredentialsStore {
	return UserCredentialsStore{TransactionStore{db: db}}
}

func (s UserCredentialsStore) GetByUserID(ctx context.Context, id uuid.UUID) (local.UserCredentials, error) {
	var c local.UserCredentials

	query := "SELECT id, password_hash FROM users WHERE id = $1"
	err := s.QuerierFromContext(ctx).QueryRow(ctx, query, id).Scan(&c.UserID, &c.PasswordHash)
	if errors.Is(err, pgx.ErrNoRows) {
		return local.UserCredentials{}, local.ErrNoCredentials
	}

	if len(c.PasswordHash) == 0 {
		// no password set (NULL in database), so return an error
		// FIXME: check NULL actually results in empty byte slice
		return local.UserCredentials{}, local.ErrNoCredentials
	}

	return c, err
}

func (s UserCredentialsStore) Save(ctx context.Context, c local.UserCredentials) error {
	query := "UPDATE users SET password_hash = $1 WHERE id = $2"
	_, err := s.QuerierFromContext(ctx).Exec(ctx, query, c.PasswordHash, c.UserID)
	return err
}
