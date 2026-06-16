package storage

import (
	"context"
	"sync"

	"gitlab.com/fynbos/mock/mockplaid/internal/models"
)

// MemoryStorage is an in-memory implementation of Storage (single process; state
// lost on restart). Used for tests and as the default when no Redis is configured.
type MemoryStorage struct {
	mu            sync.RWMutex
	linkSessions  map[string]models.LinkSession // linkToken -> session
	itemsByPublic map[string]models.Item        // publicToken -> item
	itemsByAccess map[string]models.Item        // accessToken -> item
	accountSeq    uint64
}

// NewMemoryStorage creates a new in-memory storage.
func NewMemoryStorage() Storage {
	return &MemoryStorage{
		linkSessions:  make(map[string]models.LinkSession),
		itemsByPublic: make(map[string]models.Item),
		itemsByAccess: make(map[string]models.Item),
	}
}

func (m *MemoryStorage) SaveLinkSession(_ context.Context, s models.LinkSession) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.linkSessions[s.LinkToken] = s
	return nil
}

func (m *MemoryStorage) GetLinkSession(_ context.Context, linkToken string) (models.LinkSession, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	s, ok := m.linkSessions[linkToken]
	if !ok {
		return models.LinkSession{}, ErrLinkSessionNotFound
	}
	return s, nil
}

func (m *MemoryStorage) SaveItem(_ context.Context, item models.Item) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if item.PublicToken != "" {
		m.itemsByPublic[item.PublicToken] = item
	}
	if item.AccessToken != "" {
		m.itemsByAccess[item.AccessToken] = item
	}
	return nil
}

func (m *MemoryStorage) GetItemByPublicToken(_ context.Context, publicToken string) (models.Item, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	item, ok := m.itemsByPublic[publicToken]
	if !ok {
		return models.Item{}, ErrItemNotFound
	}
	return item, nil
}

func (m *MemoryStorage) GetItemByAccessToken(_ context.Context, accessToken string) (models.Item, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	item, ok := m.itemsByAccess[accessToken]
	if !ok {
		return models.Item{}, ErrItemNotFound
	}
	return item, nil
}

func (m *MemoryStorage) DeleteItemByAccessToken(_ context.Context, accessToken string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	item, ok := m.itemsByAccess[accessToken]
	if !ok {
		// idempotent: deleting a missing item is a no-op success
		return nil
	}
	delete(m.itemsByAccess, accessToken)
	if item.PublicToken != "" {
		delete(m.itemsByPublic, item.PublicToken)
	}
	return nil
}

func (m *MemoryStorage) NextAccountSeq(_ context.Context) (uint64, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.accountSeq++
	return m.accountSeq, nil
}

func (m *MemoryStorage) Reset(_ context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.linkSessions = make(map[string]models.LinkSession)
	m.itemsByPublic = make(map[string]models.Item)
	m.itemsByAccess = make(map[string]models.Item)
	m.accountSeq = 0
	return nil
}
