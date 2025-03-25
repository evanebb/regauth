package local

import (
	"context"
	"github.com/evanebb/regauth/store"
	"github.com/google/uuid"
)

type UserCredentialsStore interface {
	store.TransactionStore
	GetByUserID(ctx context.Context, id uuid.UUID) (UserCredentials, error)
	Save(ctx context.Context, c UserCredentials) error
}
