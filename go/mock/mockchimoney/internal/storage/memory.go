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
	payments    map[string]models.Payment
	payouts     map[string]models.Payout
}

// NewMemoryStore creates an empty memory-backed store.
func NewMemoryStore() *MemoryStore {
	return &MemoryStore{
		subAccounts: make(map[string]models.SubAccount),
		payments:    make(map[string]models.Payment),
		payouts:     make(map[string]models.Payout),
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

// UpdateSubAccountKYCStatus updates KYC status for an existing sub-account.
func (s *MemoryStore) UpdateSubAccountKYCStatus(_ context.Context, id string, status string) (models.SubAccount, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	account, exists := s.subAccounts[id]
	if !exists {
		return models.SubAccount{}, ErrNotFound
	}

	account.KYCStatus = status
	s.subAccounts[id] = account
	return account, nil
}

// CreatePayment stores a payment keyed by its issue ID.
func (s *MemoryStore) CreatePayment(_ context.Context, payment models.Payment) (models.Payment, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.payments[payment.IssueID]; exists {
		return models.Payment{}, ErrAlreadyExists
	}

	s.payments[payment.IssueID] = payment
	return payment, nil
}

// GetPaymentByIssueID returns a payment by issue ID.
func (s *MemoryStore) GetPaymentByIssueID(_ context.Context, issueID string) (models.Payment, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	payment, exists := s.payments[issueID]
	if !exists {
		return models.Payment{}, ErrNotFound
	}

	return payment, nil
}

// UpdatePaymentStatus updates a payment status by issue ID.
func (s *MemoryStore) UpdatePaymentStatus(_ context.Context, issueID string, status string) (models.Payment, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	payment, exists := s.payments[issueID]
	if !exists {
		return models.Payment{}, ErrNotFound
	}

	payment.Status = status
	s.payments[issueID] = payment
	return payment, nil
}

// CreatePayout stores a payout keyed by issue ID.
func (s *MemoryStore) CreatePayout(_ context.Context, payout models.Payout) (models.Payout, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.payouts[payout.IssueID]; exists {
		return models.Payout{}, ErrAlreadyExists
	}

	s.payouts[payout.IssueID] = payout
	return payout, nil
}

// GetPayoutByChiRef returns a payout by its chiRef.
func (s *MemoryStore) GetPayoutByChiRef(_ context.Context, chiRef string) (models.Payout, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, payout := range s.payouts {
		if payout.ChiRef == chiRef {
			return payout, nil
		}
	}

	return models.Payout{}, ErrNotFound
}

// GetPayoutByIssueID returns a payout by issue ID.
func (s *MemoryStore) GetPayoutByIssueID(_ context.Context, issueID string) (models.Payout, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	payout, exists := s.payouts[issueID]
	if !exists {
		return models.Payout{}, ErrNotFound
	}

	return payout, nil
}

// UpdatePayoutStatus updates a payout status by issue ID.
func (s *MemoryStore) UpdatePayoutStatus(_ context.Context, issueID string, status string) (models.Payout, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	payout, exists := s.payouts[issueID]
	if !exists {
		return models.Payout{}, ErrNotFound
	}

	payout.Status = status
	s.payouts[issueID] = payout
	return payout, nil
}
