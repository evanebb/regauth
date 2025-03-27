package memory

import (
	"context"
	"github.com/evanebb/regauth/user"
	"github.com/google/uuid"
	"sync"
)

type UserStore struct {
	TransactionStore
	mu    sync.RWMutex
	users map[uuid.UUID]user.User
}

func NewUserStore() *UserStore {
	return &UserStore{
		users: make(map[uuid.UUID]user.User),
	}
}

func (s *UserStore) GetAll(ctx context.Context) ([]user.User, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	users := make([]user.User, 0, len(s.users))
	for _, u := range s.users {
		users = append(users, u)
	}

	return users, nil
}

func (s *UserStore) GetByID(ctx context.Context, id uuid.UUID) (user.User, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	u, ok := s.users[id]
	if !ok {
		return user.User{}, user.ErrNotFound
	}

	return u, u.IsValid()
}

func (s *UserStore) GetByUsername(ctx context.Context, username string) (user.User, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, u := range s.users {
		if string(u.Username) == username {
			return u, u.IsValid()
		}
	}

	return user.User{}, user.ErrNotFound
}

func (s *UserStore) Create(ctx context.Context, u user.User) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.users[u.ID] = u

	return nil
}

func (s *UserStore) DeleteByID(ctx context.Context, id uuid.UUID) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.users, id)

	return nil
}
