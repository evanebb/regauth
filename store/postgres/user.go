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
	db *pgxpool.Pool
}

func NewUserStore(db *pgxpool.Pool) UserStore {
	return UserStore{db: db}
}

func (s UserStore) GetAll(ctx context.Context) ([]user.User, error) {
	var users []user.User

	query := "SELECT uuid, username, firstname, lastname, role FROM users"
	rows, err := s.db.Query(ctx, query)
	if err != nil {
		return users, err
	}

	for rows.Next() {
		var u user.User

		err = rows.Scan(&u.ID, &u.Username, &u.FirstName, &u.LastName, &u.Role)
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

	query := "SELECT uuid, username, firstname, lastname, role FROM users WHERE uuid = $1"
	err := s.db.QueryRow(ctx, query, id).Scan(&u.ID, &u.Username, &u.FirstName, &u.LastName, &u.Role)
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

	query := "SELECT uuid, username, firstname, lastname, role FROM users WHERE username = $1"
	err := s.db.QueryRow(ctx, query, username).Scan(&u.ID, &u.Username, &u.FirstName, &u.LastName, &u.Role)
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

	query := "INSERT INTO users (uuid, username, firstname, lastname, role) VALUES ($1, $2, $3, $4, $5)"
	_, err = s.db.Exec(ctx, query, u.ID, u.Username, u.FirstName, u.LastName, u.Role)
	return err
}

func (s UserStore) DeleteByID(ctx context.Context, id uuid.UUID) error {
	query := "DELETE FROM users WHERE uuid = $1"
	_, err := s.db.Exec(ctx, query, id)
	return err
}
