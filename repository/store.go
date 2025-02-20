package repository

import (
	"context"
	"github.com/google/uuid"
)

type Store interface {
	GetAllByOwner(ctx context.Context, ownerId uuid.UUID) ([]Repository, error)
	GetAllPublic(ctx context.Context) ([]Repository, error)
	GetByNamespaceAndName(ctx context.Context, namespace string, name string) (Repository, error)
	GetByID(ctx context.Context, id uuid.UUID) (Repository, error)
	Create(ctx context.Context, r Repository) error
	Update(ctx context.Context, r Repository) error
	DeleteByID(ctx context.Context, id uuid.UUID) error
}
