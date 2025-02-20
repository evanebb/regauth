package postgres

import (
	"context"
	"errors"
	"github.com/evanebb/regauth/repository"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type RepositoryStore struct {
	db *pgxpool.Pool
}

func NewRepositoryStore(db *pgxpool.Pool) RepositoryStore {
	return RepositoryStore{db: db}
}

func (s RepositoryStore) GetAllByOwner(ctx context.Context, ownerId uuid.UUID) ([]repository.Repository, error) {
	var repositories []repository.Repository

	query := "SELECT uuid, namespace, name, visibility, owner_uuid FROM repositories WHERE owner_uuid = $1"
	rows, err := s.db.Query(ctx, query, ownerId)
	if err != nil {
		return repositories, err
	}

	for rows.Next() {
		var r repository.Repository

		err = rows.Scan(&r.ID, &r.Namespace, &r.Name, &r.Visibility, &r.OwnerID)
		if err != nil {
			return repositories, err
		}

		repositories = append(repositories, r)
	}

	return repositories, nil
}

func (s RepositoryStore) GetAllPublic(ctx context.Context) ([]repository.Repository, error) {
	var repositories []repository.Repository

	query := "SELECT uuid, namespace, name, visibility, owner_uuid FROM repositories WHERE visibility = 'public'"
	rows, err := s.db.Query(ctx, query)
	if err != nil {
		return repositories, err
	}

	for rows.Next() {
		var r repository.Repository

		err = rows.Scan(&r.ID, &r.Namespace, &r.Name, &r.Visibility, &r.OwnerID)
		if err != nil {
			return repositories, err
		}

		repositories = append(repositories, r)
	}

	return repositories, nil
}

func (s RepositoryStore) GetByNamespaceAndName(ctx context.Context, namespace string, name string) (repository.Repository, error) {
	var r repository.Repository

	query := "SELECT uuid, namespace, name, visibility, owner_uuid FROM repositories WHERE namespace = $1 AND name = $2"
	err := s.db.QueryRow(ctx, query, namespace, name).Scan(&r.ID, &r.Namespace, &r.Name, &r.Visibility, &r.OwnerID)
	if errors.Is(err, pgx.ErrNoRows) {
		return r, repository.ErrNotFound
	}

	return r, err
}

func (s RepositoryStore) GetByID(ctx context.Context, id uuid.UUID) (repository.Repository, error) {
	var r repository.Repository

	query := "SELECT uuid, namespace, name, visibility, owner_uuid FROM repositories WHERE uuid = $1"
	err := s.db.QueryRow(ctx, query, id).Scan(&r.ID, &r.Namespace, &r.Name, &r.Visibility, &r.OwnerID)
	if errors.Is(err, pgx.ErrNoRows) {
		return r, repository.ErrNotFound
	}

	return r, err
}

func (s RepositoryStore) Create(ctx context.Context, r repository.Repository) error {
	_, err := s.GetByID(ctx, r.ID)
	if err == nil {
		return repository.ErrAlreadyExists
	}
	if !errors.Is(err, repository.ErrNotFound) {
		return err
	}

	query := "INSERT INTO repositories (uuid, namespace, name, visibility, owner_uuid) VALUES ($1, $2, $3, $4, $5)"
	_, err = s.db.Exec(ctx, query, r.ID, r.Namespace, r.Name, r.Visibility, r.OwnerID)
	return err
}

func (s RepositoryStore) Update(ctx context.Context, r repository.Repository) error {
	_, err := s.GetByID(ctx, r.ID)
	if err != nil {
		return err
	}

	query := "UPDATE repositories SET namespace = $1, name = $2, visibility = $3, owner_uuid = $4 WHERE uuid = $5"
	_, err = s.db.Exec(ctx, query, r.Namespace, r.Name, r.Visibility, r.OwnerID, r.ID)
	return err
}

func (s RepositoryStore) DeleteByID(ctx context.Context, id uuid.UUID) error {
	query := "DELETE FROM repositories WHERE uuid = $1"
	_, err := s.db.Exec(ctx, query, id)
	return err
}
