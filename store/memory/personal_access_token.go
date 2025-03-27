package memory

import (
	"context"
	"github.com/evanebb/regauth/token"
	"github.com/google/uuid"
	"sync"
)

type PersonalAccessTokenStore struct {
	TransactionStore
	mu            sync.RWMutex
	tokens        map[string]token.PersonalAccessToken
	tokenUsageLog map[uuid.UUID][]token.UsageLogEntry
}

func NewPersonalAccessTokenStore() *PersonalAccessTokenStore {
	return &PersonalAccessTokenStore{
		tokens:        make(map[string]token.PersonalAccessToken),
		tokenUsageLog: make(map[uuid.UUID][]token.UsageLogEntry),
	}
}

func (s *PersonalAccessTokenStore) GetAllByUser(ctx context.Context, userID uuid.UUID) ([]token.PersonalAccessToken, error) {
	var tokens []token.PersonalAccessToken

	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, t := range s.tokens {
		if t.UserID == userID {
			tokens = append(tokens, t)
		}
	}

	return tokens, nil
}

func (s *PersonalAccessTokenStore) GetByID(ctx context.Context, id uuid.UUID) (token.PersonalAccessToken, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, t := range s.tokens {
		if t.ID == id {
			return t, t.IsValid()
		}
	}

	return token.PersonalAccessToken{}, token.ErrNotFound
}

func (s *PersonalAccessTokenStore) GetByPlainTextToken(ctx context.Context, plainTextToken string) (token.PersonalAccessToken, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	t, ok := s.tokens[plainTextToken]
	if !ok {
		return token.PersonalAccessToken{}, token.ErrNotFound
	}

	return t, t.IsValid()
}

func (s *PersonalAccessTokenStore) Create(ctx context.Context, t token.PersonalAccessToken, plainTextToken string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.tokens[plainTextToken] = t

	return nil
}

func (s *PersonalAccessTokenStore) DeleteByID(ctx context.Context, id uuid.UUID) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for plain, t := range s.tokens {
		if t.ID == id {
			delete(s.tokens, plain)
		}
	}

	return nil
}

func (s *PersonalAccessTokenStore) GetUsageLog(ctx context.Context, tokenID uuid.UUID) ([]token.UsageLogEntry, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	log, ok := s.tokenUsageLog[tokenID]
	if !ok {
		return make([]token.UsageLogEntry, 0), nil
	}

	return log, nil
}

func (s *PersonalAccessTokenStore) AddUsageLogEntry(ctx context.Context, e token.UsageLogEntry) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.tokenUsageLog[e.TokenID]; !ok {
		s.tokenUsageLog[e.TokenID] = make([]token.UsageLogEntry, 0)
	}

	s.tokenUsageLog[e.TokenID] = append(s.tokenUsageLog[e.TokenID], e)
	return nil
}
