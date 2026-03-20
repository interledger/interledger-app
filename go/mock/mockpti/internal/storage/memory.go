package storage

import (
	"context"
	"sync"
	"time"

	"gitlab.com/fynbos/mock/mockpti/internal/models"
)

// MemoryStorage is an in-memory implementation of the Storage interface.
type MemoryStorage struct {
	mu          sync.RWMutex
	users       map[string]*models.User
	assessments map[string][]*models.Assessment // userID -> assessments (ordered by creation)
}

// NewMemoryStorage creates a new in-memory storage.
func NewMemoryStorage() Storage {
	return &MemoryStorage{
		users:       make(map[string]*models.User),
		assessments: make(map[string][]*models.Assessment),
	}
}

// User operations

func (m *MemoryStorage) SaveUser(ctx context.Context, user *models.User) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	user.CreatedAt = time.Now()
	m.users[user.ID] = user
	return nil
}

func (m *MemoryStorage) GetUser(ctx context.Context, userID string) (*models.User, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	user, ok := m.users[userID]
	if !ok {
		return nil, ErrUserNotFound
	}
	return user, nil
}

func (m *MemoryStorage) UpdateUser(ctx context.Context, user *models.User) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, ok := m.users[user.ID]; !ok {
		return ErrUserNotFound
	}
	m.users[user.ID] = user
	return nil
}

// Assessment operations

func (m *MemoryStorage) SaveAssessment(ctx context.Context, assessment *models.Assessment) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.assessments[assessment.UserID] = append(m.assessments[assessment.UserID], assessment)
	return nil
}

func (m *MemoryStorage) GetLatestAssessment(ctx context.Context, userID string) (*models.Assessment, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	assessments, ok := m.assessments[userID]
	if !ok || len(assessments) == 0 {
		return nil, ErrAssessmentNotFound
	}
	return assessments[len(assessments)-1], nil
}

// Reset clears all data.

func (m *MemoryStorage) Reset(ctx context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.users = make(map[string]*models.User)
	m.assessments = make(map[string][]*models.Assessment)
	return nil
}
