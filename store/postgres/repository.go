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
	TransactionStore
}

func NewRepositoryStore(db *pgxpool.Pool) RepositoryStore {
	return RepositoryStore{TransactionStore{db: db}}
}

func (s RepositoryStore) GetAllByUser(ctx context.Context, userID uuid.UUID) ([]repository.Repository, error) {
	var repositories []repository.Repository

	query := `
		SELECT
			repositories.uuid,
			namespaces.name as namespace,
			repositories.name,
			repositories.visibility
		FROM repositories
		JOIN namespaces ON repositories.namespace = namespaces.uuid
		WHERE namespaces.user_uuid = $1
		OR namespaces.team_uuid IN (
			SELECT uuid
			FROM teams
			JOIN team_members ON teams.uuid = team_members.team_uuid
			WHERE team_members.user_uuid = $1
		)
		`
	rows, err := s.QuerierFromContext(ctx).Query(ctx, query, userID)
	defer rows.Close()
	if err != nil {
		return repositories, err
	}

	for rows.Next() {
		var r repository.Repository

		err = rows.Scan(&r.ID, &r.Namespace, &r.Name, &r.Visibility)
		if err != nil {
			return repositories, err
		}

		err = r.IsValid()
		if err != nil {
			return repositories, err
		}

		repositories = append(repositories, r)
	}

	return repositories, nil
}

func (s RepositoryStore) GetByNamespaceAndName(ctx context.Context, namespace string, name string) (repository.Repository, error) {
	var r repository.Repository

	query := `
		SELECT
			repositories.uuid,
			namespaces.name as namespace,
			repositories.name,
			repositories.visibility
		FROM repositories
		JOIN namespaces ON repositories.namespace = namespaces.uuid
		WHERE namespaces.name = $1 AND repositories.name = $2
		`
	err := s.QuerierFromContext(ctx).QueryRow(ctx, query, namespace, name).Scan(&r.ID, &r.Namespace, &r.Name, &r.Visibility)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return r, repository.ErrNotFound
		}

		return r, err
	}

	return r, r.IsValid()
}

func (s RepositoryStore) GetByID(ctx context.Context, id uuid.UUID) (repository.Repository, error) {
	var r repository.Repository

	query := `
		SELECT
			repositories.uuid,
			namespaces.name as namespace,
			repositories.name,
			repositories.visibility
		FROM repositories
		JOIN namespaces ON repositories.namespace = namespaces.uuid
		WHERE repositories.uuid = $1
		`
	err := s.QuerierFromContext(ctx).QueryRow(ctx, query, id).Scan(&r.ID, &r.Namespace, &r.Name, &r.Visibility)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return r, repository.ErrNotFound
		}

		return r, err
	}

	return r, r.IsValid()
}

func (s RepositoryStore) Create(ctx context.Context, r repository.Repository) error {
	_, err := s.GetByID(ctx, r.ID)
	if err == nil {
		return repository.ErrAlreadyExists
	}
	if !errors.Is(err, repository.ErrNotFound) {
		return err
	}

	query := `
		INSERT INTO repositories (uuid, namespace, name, visibility)
		SELECT $1, uuid, $2, $3
		FROM namespaces
		WHERE name = $4
		`

	_, err = s.QuerierFromContext(ctx).Exec(ctx, query, r.ID, r.Name, r.Visibility, r.Namespace)
	return err
}

func (s RepositoryStore) DeleteByID(ctx context.Context, id uuid.UUID) error {
	query := "DELETE FROM repositories WHERE uuid = $1"
	_, err := s.QuerierFromContext(ctx).Exec(ctx, query, id)
	return err
}
