package pat

import (
	"context"
	"github.com/google/uuid"
)

type Store interface {
	GetAllForUser(ctx context.Context, userID uuid.UUID) ([]PersonalAccessToken, error)
	GetByID(ctx context.Context, id uuid.UUID) (PersonalAccessToken, error)
	Create(ctx context.Context, t PersonalAccessToken) error
	DeleteByID(ctx context.Context, id uuid.UUID) error
	GetUsageLog(ctx context.Context, tokenID uuid.UUID) ([]UsageLogEntry, error)
	AddUsageLogEntry(ctx context.Context, e UsageLogEntry) error
}
