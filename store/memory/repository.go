package memory

import (
	"context"
	"github.com/evanebb/regauth/repository"
	"github.com/google/uuid"
	"slices"
	"sync"
)

type RepositoryStore struct {
	TransactionStore
	mu           sync.RWMutex
	repositories map[uuid.UUID]repository.Repository
}

func NewRepositoryStore() *RepositoryStore {
	return &RepositoryStore{
		repositories: make(map[uuid.UUID]repository.Repository),
	}
}

func (s *RepositoryStore) GetAllByNamespace(ctx context.Context, namespaces ...string) ([]repository.Repository, error) {
	var repositories []repository.Repository

	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, r := range s.repositories {
		if slices.Contains(namespaces, r.Namespace) {
			repositories = append(repositories, r)
		}
	}

	return repositories, nil
}

func (s *RepositoryStore) GetByNamespaceAndName(ctx context.Context, namespace string, name string) (repository.Repository, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, r := range s.repositories {
		if r.Namespace == namespace && string(r.Name) == name {
			return r, r.IsValid()
		}
	}

	return repository.Repository{}, repository.ErrNotFound
}

func (s *RepositoryStore) GetByID(ctx context.Context, id uuid.UUID) (repository.Repository, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	r, ok := s.repositories[id]
	if !ok {
		return repository.Repository{}, repository.ErrNotFound
	}

	return r, r.IsValid()
}

func (s *RepositoryStore) Create(ctx context.Context, r repository.Repository) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.repositories[r.ID] = r

	return nil
}

func (s *RepositoryStore) DeleteByID(ctx context.Context, id uuid.UUID) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.repositories, id)

	return nil
}
