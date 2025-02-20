package user

import (
	"context"
	"github.com/google/uuid"
)

type Store interface {
	GetAll(ctx context.Context) ([]User, error)
	GetByID(ctx context.Context, id uuid.UUID) (User, error)
	GetByUsername(ctx context.Context, username string) (User, error)
	Create(ctx context.Context, u User) error
	DeleteByID(ctx context.Context, id uuid.UUID) error
}
