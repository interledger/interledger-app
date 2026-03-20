package storage

import (
	"context"
	"sync"

	"gitlab.com/fynbos/mock/mockchimoney/internal/models"
)

// MemoryStore is an in-memory Store implementation guarded by a mutex.
type MemoryStore struct {
	mu          sync.RWMutex
	subAccounts map[string]models.SubAccount
}

// NewMemoryStore creates an empty memory-backed store.
func NewMemoryStore() *MemoryStore {
	return &MemoryStore{
		subAccounts: make(map[string]models.SubAccount),
	}
}

// CreateSubAccount stores a sub-account keyed by its ID.
func (s *MemoryStore) CreateSubAccount(_ context.Context, account models.SubAccount) (models.SubAccount, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.subAccounts[account.ID]; exists {
		return models.SubAccount{}, ErrAlreadyExists
	}

	s.subAccounts[account.ID] = account
	return account, nil
}

// GetSubAccount loads a sub-account by ID.
func (s *MemoryStore) GetSubAccount(_ context.Context, id string) (models.SubAccount, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	account, exists := s.subAccounts[id]
	if !exists {
		return models.SubAccount{}, ErrNotFound
	}

	return account, nil
}

// ListSubAccounts returns all known sub-accounts.
func (s *MemoryStore) ListSubAccounts(_ context.Context) ([]models.SubAccount, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	accounts := make([]models.SubAccount, 0, len(s.subAccounts))
	for _, account := range s.subAccounts {
		accounts = append(accounts, account)
	}

	return accounts, nil
}
