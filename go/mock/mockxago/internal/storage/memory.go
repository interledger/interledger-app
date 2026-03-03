package storage

import (
	"context"
	"sync"
	"time"

	"gitlab.com/fynbos/mock/mockxago/internal/models"
)

// MemoryStorage is an in-memory implementation of the Storage interface
type MemoryStorage struct {
	mu            sync.RWMutex
	tokens        map[string]*models.AccessToken
	tokenAccounts map[string]string
}

// NewMemoryStorage creates a new in-memory storage
func NewMemoryStorage() Storage {
	return &MemoryStorage{
		tokens:        make(map[string]*models.AccessToken),
		tokenAccounts: make(map[string]string),
	}
}

// Token operations

func (m *MemoryStorage) SaveAccessToken(ctx context.Context, token *models.AccessToken) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	token.CreatedAt = time.Now()
	m.tokens[token.Token] = token
	return nil
}

func (m *MemoryStorage) GetAccessToken(ctx context.Context, tokenValue string) (*models.AccessToken, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	token, ok := m.tokens[tokenValue]
	if !ok {
		return nil, ErrTokenNotFound
	}

	if token.IsExpired() {
		return nil, ErrTokenExpired
	}

	return token, nil
}

func (m *MemoryStorage) InvalidateAccessToken(ctx context.Context, tokenValue string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	delete(m.tokens, tokenValue)
	return nil
}

func (m *MemoryStorage) SaveTokenAccount(ctx context.Context, tokenValue string, accountID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if tokenValue == "" || accountID == "" {
		return nil
	}

	m.tokenAccounts[tokenValue] = accountID
	return nil
}

func (m *MemoryStorage) GetAccountIDByToken(ctx context.Context, tokenValue string) (string, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	accountID, ok := m.tokenAccounts[tokenValue]
	if !ok {
		return "", ErrTokenNotFound
	}

	return accountID, nil
}

// Reset clears all data in storage (used for test state reset between scenarios)
func (m *MemoryStorage) Reset(ctx context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.tokens = make(map[string]*models.AccessToken)
	m.tokenAccounts = make(map[string]string)

	return nil
}
