package postgres

import (
	"context"
	"errors"
	"github.com/evanebb/regauth/user"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type UserStore struct {
	TransactionStore
}

func NewUserStore(db *pgxpool.Pool) UserStore {
	return UserStore{TransactionStore{db: db}}
}

func (s UserStore) GetAll(ctx context.Context) ([]user.User, error) {
	var users []user.User

	query := "SELECT uuid, username, role FROM users"
	rows, err := s.QuerierFromContext(ctx).Query(ctx, query)
	defer rows.Close()
	if err != nil {
		return users, err
	}

	for rows.Next() {
		var u user.User

		err = rows.Scan(&u.ID, &u.Username, &u.Role)
		if err != nil {
			return users, err
		}

		err = u.IsValid()
		if err != nil {
			return users, err
		}

		users = append(users, u)
	}

	return users, nil
}

func (s UserStore) GetByID(ctx context.Context, id uuid.UUID) (user.User, error) {
	var u user.User

	query := "SELECT uuid, username, role FROM users WHERE uuid = $1"
	err := s.QuerierFromContext(ctx).QueryRow(ctx, query, id).Scan(&u.ID, &u.Username, &u.Role)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return u, user.ErrNotFound
		}

		return u, err
	}

	return u, u.IsValid()
}

func (s UserStore) GetByUsername(ctx context.Context, username string) (user.User, error) {
	var u user.User

	query := "SELECT uuid, username, role FROM users WHERE username = $1"
	err := s.QuerierFromContext(ctx).QueryRow(ctx, query, username).Scan(&u.ID, &u.Username, &u.Role)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return u, user.ErrNotFound
		}

		return u, err
	}

	return u, u.IsValid()
}

func (s UserStore) Create(ctx context.Context, u user.User) error {
	_, err := s.GetByID(ctx, u.ID)
	if err == nil {
		return user.ErrAlreadyExists
	}
	if !errors.Is(err, user.ErrNotFound) {
		return err
	}

	tx, err := s.QuerierFromContext(ctx).Begin(ctx)
	if err != nil {
		return err
	}

	query := "INSERT INTO users (uuid, username, role) VALUES ($1, $2, $3)"
	if _, err = tx.Exec(ctx, query, u.ID, u.Username, u.Role); err != nil {
		_ = tx.Rollback(ctx)
		return err
	}

	nsQuery := "INSERT INTO namespaces (uuid, name, user_uuid) VALUES ($1, $2, $3)"
	if _, err := tx.Exec(ctx, nsQuery, uuid.New(), u.Username, u.ID); err != nil {
		_ = tx.Rollback(ctx)
		return err
	}

	_ = tx.Commit(ctx)
	return nil
}

func (s UserStore) DeleteByID(ctx context.Context, id uuid.UUID) error {
	query := "DELETE FROM users WHERE uuid = $1"
	_, err := s.QuerierFromContext(ctx).Exec(ctx, query, id)
	return err
}
