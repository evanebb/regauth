package repository

import (
	"context"
	"github.com/evanebb/regauth/store"
	"github.com/google/uuid"
)

type Store interface {
	store.TransactionStore
	GetAllByNamespace(ctx context.Context, namespaces ...string) ([]Repository, error)
	GetByNamespaceAndName(ctx context.Context, namespace string, name string) (Repository, error)
	GetByID(ctx context.Context, id uuid.UUID) (Repository, error)
	Create(ctx context.Context, r Repository) error
	DeleteByID(ctx context.Context, id uuid.UUID) error
}
