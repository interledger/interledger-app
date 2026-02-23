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
	tokenAccounts         map[string]string
	subAccounts           map[string]*models.SubAccount
	subAccountsByWallet   map[string]*models.SubAccount
	beneficiaries         map[string]*models.Beneficiary
	beneficiariesByWallet map[string][]*models.Beneficiary
	transactions          map[string]*models.Transaction
	idempotencyKeys       map[string]string
	balances              map[string]map[string]balanceEntry // [walletID][currency] -> entry
	deposits              map[string]*models.Deposit
	depositsByReference   map[string]*models.Deposit
	jobs                  map[string]*models.Job
}

type balanceEntry struct {
	Available float64
	Reserved  float64
}

// NewMemoryStorage creates a new in-memory storage
func NewMemoryStorage() Storage {
	return &MemoryStorage{
		tokens:                make(map[string]*models.AccessToken),
		tokenAccounts:         make(map[string]string),
		subAccounts:           make(map[string]*models.SubAccount),
		subAccountsByWallet:   make(map[string]*models.SubAccount),
		beneficiaries:         make(map[string]*models.Beneficiary),
		beneficiariesByWallet: make(map[string][]*models.Beneficiary),
		transactions:          make(map[string]*models.Transaction),
		idempotencyKeys:       make(map[string]string),
		balances:              make(map[string]map[string]balanceEntry),
		deposits:              make(map[string]*models.Deposit),
		depositsByReference:   make(map[string]*models.Deposit),
		jobs:                  make(map[string]*models.Job),
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
	// Index by AccountID for querying beneficiaries associated with a specific account
	m.beneficiariesByWallet[beneficiary.AccountID] = append(m.beneficiariesByWallet[beneficiary.AccountID], beneficiary)
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

func (m *MemoryStorage) ListTransactionsByAccount(ctx context.Context, accountID string, limit int, offset int) ([]*models.Transaction, int, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var transactions []*models.Transaction
	for _, tx := range m.transactions {
		if tx.AccountID == accountID {
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

func (m *MemoryStorage) GetBalance(ctx context.Context, walletID string, currency string) (float64, float64, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if walletBalances, ok := m.balances[walletID]; ok {
		entry := walletBalances[currency]
		return entry.Available, entry.Reserved, nil
	}

	return 0.0, 0.0, nil
}

func (m *MemoryStorage) SetBalance(ctx context.Context, walletID string, currency string, available float64, reserved float64) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, ok := m.balances[walletID]; !ok {
		m.balances[walletID] = make(map[string]balanceEntry)
	}

	m.balances[walletID][currency] = balanceEntry{Available: available, Reserved: reserved}
	return nil
}

func (m *MemoryStorage) AddBalance(ctx context.Context, walletID string, currency string, amount float64) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, ok := m.balances[walletID]; !ok {
		m.balances[walletID] = make(map[string]balanceEntry)
	}

	entry := m.balances[walletID][currency]
	entry.Available += amount
	m.balances[walletID][currency] = entry
	return nil
}

func (m *MemoryStorage) SubtractBalance(ctx context.Context, walletID string, currency string, amount float64) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, ok := m.balances[walletID]; !ok {
		m.balances[walletID] = make(map[string]balanceEntry)
	}

	entry := m.balances[walletID][currency]
	if entry.Available < amount {
		return ErrInsufficientBalance
	}

	entry.Available -= amount
	m.balances[walletID][currency] = entry
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

// ClearDeposits removes all deposit records (used for test state reset between scenarios).
func (m *MemoryStorage) ClearDeposits(ctx context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.deposits = make(map[string]*models.Deposit)
	m.depositsByReference = make(map[string]*models.Deposit)
	return nil
}

// ClearBalances removes all balance records (used for test state reset between scenarios).
func (m *MemoryStorage) ClearBalances(ctx context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.balances = make(map[string]map[string]balanceEntry)
	return nil
}

// Job operations

// SaveJob saves a job to storage
func (m *MemoryStorage) SaveJob(ctx context.Context, job *models.Job) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.jobs[job.ID] = job
	return nil
}

// GetJob retrieves a job by ID
func (m *MemoryStorage) GetJob(ctx context.Context, jobID string) (*models.Job, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	job, ok := m.jobs[jobID]
	if !ok {
		return nil, nil // Not found
	}
	return job, nil
}

// ListReadyJobs returns jobs that are ready for processing (status=pending and NotBefore <= now)
func (m *MemoryStorage) ListReadyJobs(ctx context.Context, limit int) ([]*models.Job, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	now := time.Now()
	var ready []*models.Job

	for _, job := range m.jobs {
		if job.Status == "pending" && !job.NotBefore.After(now) {
			ready = append(ready, job)
		}
	}

	// Sort by NotBefore ascending (oldest first)
	// Simple bubble sort for small lists
	for i := 0; i < len(ready); i++ {
		for j := i + 1; j < len(ready); j++ {
			if ready[i].NotBefore.After(ready[j].NotBefore) {
				ready[i], ready[j] = ready[j], ready[i]
			}
		}
	}

	if len(ready) > limit {
		ready = ready[:limit]
	}

	return ready, nil
}

// UpdateJobStatus updates a job's status, completed_at, and last_error
func (m *MemoryStorage) UpdateJobStatus(ctx context.Context, jobID string, status string, completedAt *time.Time, lastError string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	job, ok := m.jobs[jobID]
	if !ok {
		return nil // Job not found, ignore
	}

	job.Status = status
	job.CompletedAt = completedAt
	job.LastError = lastError
	return nil
}

// IncrementJobAttempts increments the attempts counter for a job
func (m *MemoryStorage) IncrementJobAttempts(ctx context.Context, jobID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	job, ok := m.jobs[jobID]
	if !ok {
		return nil // Job not found, ignore
	}

	job.Attempts++
	return nil
}

// ClearJobs removes all job records (used for test state reset between scenarios)
func (m *MemoryStorage) ClearJobs(ctx context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.jobs = make(map[string]*models.Job)
	return nil
}

// Reset clears all data in storage (used for test state reset between scenarios)
func (m *MemoryStorage) Reset(ctx context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.tokens = make(map[string]*models.AccessToken)
	m.tokenAccounts = make(map[string]string)
	m.subAccounts = make(map[string]*models.SubAccount)
	m.subAccountsByWallet = make(map[string]*models.SubAccount)
	m.beneficiaries = make(map[string]*models.Beneficiary)
	m.beneficiariesByWallet = make(map[string][]*models.Beneficiary)
	m.transactions = make(map[string]*models.Transaction)
	m.idempotencyKeys = make(map[string]string)
	m.balances = make(map[string]map[string]balanceEntry)
	m.deposits = make(map[string]*models.Deposit)
	m.depositsByReference = make(map[string]*models.Deposit)
	m.jobs = make(map[string]*models.Job)

	return nil
}
