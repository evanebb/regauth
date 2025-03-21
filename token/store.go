package token

import (
	"context"
	"github.com/evanebb/regauth/store"
	"github.com/google/uuid"
)

type Store interface {
	store.TransactionStore
	GetAllByUser(ctx context.Context, userID uuid.UUID) ([]PersonalAccessToken, error)
	GetByID(ctx context.Context, id uuid.UUID) (PersonalAccessToken, error)
	GetByPlainTextToken(ctx context.Context, plainTextToken string) (PersonalAccessToken, error)
	// Create will create the given token in the underlying store. Note that the plain-text token is a password, and
	// must be hashed by the implementor before storing it.
	Create(ctx context.Context, t PersonalAccessToken, plainTextToken string) error
	DeleteByID(ctx context.Context, id uuid.UUID) error
	GetUsageLog(ctx context.Context, tokenID uuid.UUID) ([]UsageLogEntry, error)
	AddUsageLogEntry(ctx context.Context, e UsageLogEntry) error
}
