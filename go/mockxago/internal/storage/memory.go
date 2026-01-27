package storage

import (
	"context"
	"sync"
	"time"

	"gitlab.com/fynbos/mockxago/internal/models"
)

// MemoryStorage is an in-memory implementation of the Storage interface
type MemoryStorage struct {
	mu                    sync.RWMutex
	tokens                map[string]*models.AccessToken
	subAccounts           map[string]*models.SubAccount
	subAccountsByWallet   map[string]*models.SubAccount
	beneficiaries         map[string]*models.Beneficiary
	beneficiariesByWallet map[string][]*models.Beneficiary
	transactions          map[string]*models.Transaction
	idempotencyKeys       map[string]string
	balances              map[string]map[string]float64 // [walletID][currency] -> amount
	deposits              map[string]*models.Deposit
	depositsByReference   map[string]*models.Deposit
}

// NewMemoryStorage creates a new in-memory storage
func NewMemoryStorage() Storage {
	return &MemoryStorage{
		tokens:                make(map[string]*models.AccessToken),
		subAccounts:           make(map[string]*models.SubAccount),
		subAccountsByWallet:   make(map[string]*models.SubAccount),
		beneficiaries:         make(map[string]*models.Beneficiary),
		beneficiariesByWallet: make(map[string][]*models.Beneficiary),
		transactions:          make(map[string]*models.Transaction),
		idempotencyKeys:       make(map[string]string),
		balances:              make(map[string]map[string]float64),
		deposits:              make(map[string]*models.Deposit),
		depositsByReference:   make(map[string]*models.Deposit),
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

// Sub-account operations

func (m *MemoryStorage) SaveSubAccount(ctx context.Context, account *models.SubAccount) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	account.CreatedAt = time.Now()
	account.UpdatedAt = time.Now()
	m.subAccounts[account.AccountID] = account
	m.subAccountsByWallet[account.WalletID] = account
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

	if _, ok := m.subAccounts[account.AccountID]; !ok {
		return ErrSubAccountNotFound
	}

	account.UpdatedAt = time.Now()
	m.subAccounts[account.AccountID] = account
	m.subAccountsByWallet[account.WalletID] = account
	return nil
}

// Beneficiary operations

func (m *MemoryStorage) SaveBeneficiary(ctx context.Context, beneficiary *models.Beneficiary) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	beneficiary.CreatedAt = time.Now()
	beneficiary.UpdatedAt = time.Now()
	m.beneficiaries[beneficiary.ID] = beneficiary
	m.beneficiariesByWallet[beneficiary.WalletID] = append(m.beneficiariesByWallet[beneficiary.WalletID], beneficiary)
	return nil
}

func (m *MemoryStorage) GetBeneficiary(ctx context.Context, beneficiaryID string) (*models.Beneficiary, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	beneficiary, ok := m.beneficiaries[beneficiaryID]
	if !ok {
		return nil, ErrBeneficiaryNotFound
	}

	return beneficiary, nil
}

func (m *MemoryStorage) ListBeneficiariesByWallet(ctx context.Context, walletID string, limit int, offset int) ([]*models.Beneficiary, int, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	beneficiaries, ok := m.beneficiariesByWallet[walletID]
	if !ok {
		return []*models.Beneficiary{}, 0, nil
	}

	total := len(beneficiaries)
	end := offset + limit
	if end > total {
		end = total
	}

	if offset >= total {
		return []*models.Beneficiary{}, total, nil
	}

	return beneficiaries[offset:end], total, nil
}

func (m *MemoryStorage) UpdateBeneficiaryStatus(ctx context.Context, beneficiaryID string, status string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	beneficiary, ok := m.beneficiaries[beneficiaryID]
	if !ok {
		return ErrBeneficiaryNotFound
	}

	beneficiary.Status = status
	beneficiary.UpdatedAt = time.Now()
	m.beneficiaries[beneficiaryID] = beneficiary

	// Update in wallet list
	if benList, ok := m.beneficiariesByWallet[beneficiary.WalletID]; ok {
		for i, b := range benList {
			if b.ID == beneficiaryID {
				benList[i] = beneficiary
				break
			}
		}
	}

	return nil
}

// Transaction operations

func (m *MemoryStorage) SaveTransaction(ctx context.Context, transaction *models.Transaction) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	transaction.CreatedAt = time.Now()
	m.transactions[transaction.ID] = transaction
	return nil
}

func (m *MemoryStorage) GetTransaction(ctx context.Context, transactionID string) (*models.Transaction, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	transaction, ok := m.transactions[transactionID]
	if !ok {
		return nil, ErrTokenNotFound
	}

	return transaction, nil
}

func (m *MemoryStorage) GetTransactionByIdempotencyKey(ctx context.Context, key string) (*models.Transaction, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	txID, ok := m.idempotencyKeys[key]
	if !ok {
		return nil, ErrTokenNotFound
	}

	transaction, ok := m.transactions[txID]
	if !ok {
		return nil, ErrTokenNotFound
	}

	return transaction, nil
}

func (m *MemoryStorage) SaveIdempotencyKey(ctx context.Context, key string, transactionID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.idempotencyKeys[key] = transactionID
	return nil
}

func (m *MemoryStorage) ListTransactionsByWallet(ctx context.Context, walletID string, limit int, offset int) ([]*models.Transaction, int, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var transactions []*models.Transaction
	for _, tx := range m.transactions {
		if tx.WalletID == walletID {
			transactions = append(transactions, tx)
		}
	}

	total := len(transactions)
	end := offset + limit
	if end > total {
		end = total
	}

	if offset >= total {
		return []*models.Transaction{}, total, nil
	}

	return transactions[offset:end], total, nil
}

func (m *MemoryStorage) UpdateTransactionStatus(ctx context.Context, transactionID string, status string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	transaction, ok := m.transactions[transactionID]
	if !ok {
		return ErrTokenNotFound
	}

	transaction.Status = status
	if status == "completed" {
		now := time.Now()
		transaction.SettledAt = &now
	}
	m.transactions[transactionID] = transaction
	return nil
}

// Balance operations

func (m *MemoryStorage) GetBalance(ctx context.Context, walletID string, currency string) (float64, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if walletBalances, ok := m.balances[walletID]; ok {
		return walletBalances[currency], nil
	}

	return 0.0, nil
}

func (m *MemoryStorage) SetBalance(ctx context.Context, walletID string, currency string, amount float64) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, ok := m.balances[walletID]; !ok {
		m.balances[walletID] = make(map[string]float64)
	}

	m.balances[walletID][currency] = amount
	return nil
}

func (m *MemoryStorage) AddBalance(ctx context.Context, walletID string, currency string, amount float64) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, ok := m.balances[walletID]; !ok {
		m.balances[walletID] = make(map[string]float64)
	}

	m.balances[walletID][currency] += amount
	return nil
}

func (m *MemoryStorage) SubtractBalance(ctx context.Context, walletID string, currency string, amount float64) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, ok := m.balances[walletID]; !ok {
		m.balances[walletID] = make(map[string]float64)
	}

	if m.balances[walletID][currency] < amount {
		return ErrInsufficientBalance
	}

	m.balances[walletID][currency] -= amount
	return nil
}

// Deposit operations

func (m *MemoryStorage) SaveDeposit(ctx context.Context, deposit *models.Deposit) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	deposit.CreatedAt = time.Now()
	m.deposits[deposit.ID] = deposit
	m.depositsByReference[deposit.DepositReference] = deposit
	return nil
}

func (m *MemoryStorage) GetDeposit(ctx context.Context, depositID string) (*models.Deposit, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	deposit, ok := m.deposits[depositID]
	if !ok {
		return nil, ErrTokenNotFound
	}

	return deposit, nil
}

func (m *MemoryStorage) GetDepositByReference(ctx context.Context, reference string) (*models.Deposit, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	deposit, ok := m.depositsByReference[reference]
	if !ok {
		return nil, ErrTokenNotFound
	}

	return deposit, nil
}

func (m *MemoryStorage) ListDeposits(ctx context.Context, limit int, offset int) ([]*models.Deposit, int, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var deposits []*models.Deposit
	for _, d := range m.deposits {
		deposits = append(deposits, d)
	}

	total := len(deposits)
	end := offset + limit
	if end > total {
		end = total
	}

	if offset >= total {
		return []*models.Deposit{}, total, nil
	}

	return deposits[offset:end], total, nil
}

func (m *MemoryStorage) UpdateDepositStatus(ctx context.Context, depositID string, status string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	deposit, ok := m.deposits[depositID]
	if !ok {
		return ErrTokenNotFound
	}

	deposit.Status = status
	if status == "completed" {
		now := time.Now()
		deposit.SettledAt = &now
	}
	m.deposits[depositID] = deposit
	return nil
}
