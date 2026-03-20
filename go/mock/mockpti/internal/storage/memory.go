package storage

import (
	"context"
	"sync"
	"time"

	"gitlab.com/fynbos/mock/mockpti/internal/models"
)

// MemoryStorage is an in-memory implementation of the Storage interface.
type MemoryStorage struct {
	mu                  sync.RWMutex
	users               map[string]*models.User
	assessments         map[string][]*models.Assessment                  // userID -> assessments (ordered by creation)
	wallets             map[string]map[string]*models.Wallet             // userID -> walletID -> wallet
	paymentInformations map[string]map[string]*models.PaymentInformation // userID -> piID -> payment info
}

// NewMemoryStorage creates a new in-memory storage.
func NewMemoryStorage() Storage {
	return &MemoryStorage{
		users:               make(map[string]*models.User),
		assessments:         make(map[string][]*models.Assessment),
		wallets:             make(map[string]map[string]*models.Wallet),
		paymentInformations: make(map[string]map[string]*models.PaymentInformation),
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
	m.wallets = make(map[string]map[string]*models.Wallet)
	m.paymentInformations = make(map[string]map[string]*models.PaymentInformation)
	return nil
}

// Wallet operations

func (m *MemoryStorage) SaveWallet(ctx context.Context, wallet *models.Wallet) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.wallets[wallet.UserID] == nil {
		m.wallets[wallet.UserID] = make(map[string]*models.Wallet)
	}
	m.wallets[wallet.UserID][wallet.WalletID] = wallet
	return nil
}

func (m *MemoryStorage) GetWallet(ctx context.Context, userID, walletID string) (*models.Wallet, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	userWallets, ok := m.wallets[userID]
	if !ok {
		return nil, ErrWalletNotFound
	}
	wallet, ok := userWallets[walletID]
	if !ok {
		return nil, ErrWalletNotFound
	}
	return wallet, nil
}

func (m *MemoryStorage) ListWallets(ctx context.Context, userID string) ([]*models.Wallet, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	userWallets := m.wallets[userID]
	result := make([]*models.Wallet, 0, len(userWallets))
	for _, w := range userWallets {
		result = append(result, w)
	}
	return result, nil
}

// Payment information operations

func (m *MemoryStorage) SavePaymentInformation(ctx context.Context, pi *models.PaymentInformation) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.paymentInformations[pi.UserID] == nil {
		m.paymentInformations[pi.UserID] = make(map[string]*models.PaymentInformation)
	}
	m.paymentInformations[pi.UserID][pi.ID] = pi
	return nil
}

func (m *MemoryStorage) GetPaymentInformation(ctx context.Context, userID, piID string) (*models.PaymentInformation, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	userPIs, ok := m.paymentInformations[userID]
	if !ok {
		return nil, ErrPaymentInformationNotFound
	}
	pi, ok := userPIs[piID]
	if !ok {
		return nil, ErrPaymentInformationNotFound
	}
	return pi, nil
}
