package storage

import (
	"context"
	"sync"
	"time"

	"gitlab.com/fynbos/mock/mockxago/internal/models"
)

// MemoryStorage is an in-memory implementation of the Storage interface
type MemoryStorage struct {
	mu                  sync.RWMutex
	tokens              map[string]*models.AccessToken
	tokenAccounts       map[string]string
	subAccounts         map[string]*models.SubAccount
	subAccountsByWallet map[string]*models.SubAccount
}

// NewMemoryStorage creates a new in-memory storage
func NewMemoryStorage() Storage {
	return &MemoryStorage{
		tokens:              make(map[string]*models.AccessToken),
		tokenAccounts:       make(map[string]string),
		subAccounts:         make(map[string]*models.SubAccount),
		subAccountsByWallet: make(map[string]*models.SubAccount),
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
	m.subAccounts = make(map[string]*models.SubAccount)
	m.subAccountsByWallet = make(map[string]*models.SubAccount)

	return nil
}

// Sub-account operations

func (m *MemoryStorage) SaveSubAccount(ctx context.Context, account *models.SubAccount) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	now := time.Now()
	account.CreatedAt = now
	account.UpdatedAt = now
	m.subAccounts[account.AccountID] = account
	if account.WalletID != "" {
		m.subAccountsByWallet[account.WalletID] = account
	}
	return nil
}

func (m *MemoryStorage) GetSubAccount(ctx context.Context, accountID string) (*models.SubAccount, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	account, ok := m.subAccounts[accountID]
	if !ok {
		return nil, ErrSubAccountNotFound
	}
	return account, nil
}

func (m *MemoryStorage) GetSubAccountByWalletID(ctx context.Context, walletID string) (*models.SubAccount, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	account, ok := m.subAccountsByWallet[walletID]
	if !ok {
		return nil, ErrSubAccountNotFound
	}
	return account, nil
}

func (m *MemoryStorage) UpdateSubAccount(ctx context.Context, account *models.SubAccount) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	account.UpdatedAt = time.Now()
	m.subAccounts[account.AccountID] = account
	if account.WalletID != "" {
		m.subAccountsByWallet[account.WalletID] = account
	}
	return nil
}
