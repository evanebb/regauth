package user

import (
	"context"
	"github.com/evanebb/regauth/store"
	"github.com/google/uuid"
)

type Store interface {
	store.TransactionStore
	GetAll(ctx context.Context) ([]User, error)
	GetByID(ctx context.Context, id uuid.UUID) (User, error)
	GetByUsername(ctx context.Context, username string) (User, error)
	Create(ctx context.Context, u User) error
	DeleteByID(ctx context.Context, id uuid.UUID) error
}

type TeamStore interface {
	store.TransactionStore
	GetAllByUser(ctx context.Context, userID uuid.UUID) ([]Team, error)
	GetByID(ctx context.Context, id uuid.UUID) (Team, error)
	GetByName(ctx context.Context, name string) (Team, error)
	Create(ctx context.Context, t Team) error
	DeleteByID(ctx context.Context, id uuid.UUID) error
	GetTeamMember(ctx context.Context, teamID uuid.UUID, userID uuid.UUID) (TeamMember, error)
	GetTeamMembers(ctx context.Context, teamID uuid.UUID) ([]TeamMember, error)
	AddTeamMember(ctx context.Context, m TeamMember) error
	RemoveTeamMember(ctx context.Context, teamID uuid.UUID, userID uuid.UUID) error
}
