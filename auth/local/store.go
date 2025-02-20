package local

import (
	"context"
	"github.com/google/uuid"
)

type AuthUserStore interface {
	GetByID(ctx context.Context, id uuid.UUID) (AuthUser, error)
	GetByUsername(ctx context.Context, username string) (AuthUser, error)
	Update(ctx context.Context, u AuthUser) error
	Create(ctx context.Context, u AuthUser) error
	DeleteByID(ctx context.Context, id uuid.UUID) error
}
