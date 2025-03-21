package local

import (
	"context"
	"github.com/evanebb/regauth/store"
	"github.com/google/uuid"
)

type AuthUserStore interface {
	store.TransactionStore
	GetByID(ctx context.Context, id uuid.UUID) (AuthUser, error)
	GetByUsername(ctx context.Context, username string) (AuthUser, error)
	Update(ctx context.Context, u AuthUser) error
	Create(ctx context.Context, u AuthUser) error
	DeleteByID(ctx context.Context, id uuid.UUID) error
}
