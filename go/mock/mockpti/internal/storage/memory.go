package storage

import (
	"context"
	"sort"
	"sync"
	"time"

	"github.com/interledger/interledger-app/go/mock/mockpti/internal/models"
)

// MemoryStorage is an in-memory implementation of the Storage interface.
type MemoryStorage struct {
	mu                  sync.RWMutex
	users               map[string]*models.User
	assessments         map[string][]*models.Assessment                  // userID -> assessments (ordered by creation)
	wallets             map[string]map[string]*models.Wallet             // userID -> walletID -> wallet
	paymentInformations map[string]map[string]*models.PaymentInformation // userID -> piID -> payment info
	transactions        map[string]*models.Transaction                   // requestID -> transaction
	transactionUpdates  map[string][]*models.TransactionUpdate           // requestID -> updates (ordered)
	jobs                map[string]*models.Job                           // jobID -> job
}

// NewMemoryStorage creates a new in-memory storage.
func NewMemoryStorage() Storage {
	return &MemoryStorage{
		users:               make(map[string]*models.User),
		assessments:         make(map[string][]*models.Assessment),
		wallets:             make(map[string]map[string]*models.Wallet),
		paymentInformations: make(map[string]map[string]*models.PaymentInformation),
		transactions:        make(map[string]*models.Transaction),
		transactionUpdates:  make(map[string][]*models.TransactionUpdate),
		jobs:                make(map[string]*models.Job),
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
	m.transactions = make(map[string]*models.Transaction)
	m.transactionUpdates = make(map[string][]*models.TransactionUpdate)
	m.jobs = make(map[string]*models.Job)
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

// Transaction operations

func (m *MemoryStorage) SaveTransaction(ctx context.Context, tx *models.Transaction) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	tx.CreatedAt = time.Now()
	m.transactions[tx.RequestID] = tx
	return nil
}

func (m *MemoryStorage) GetTransaction(ctx context.Context, requestID string) (*models.Transaction, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	tx, ok := m.transactions[requestID]
	if !ok {
		return nil, ErrTransactionNotFound
	}
	return tx, nil
}

func (m *MemoryStorage) SaveTransactionUpdate(ctx context.Context, update *models.TransactionUpdate) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.transactionUpdates[update.RequestID] = append(m.transactionUpdates[update.RequestID], update)
	return nil
}

// Job operations

func (m *MemoryStorage) SaveJob(ctx context.Context, job *models.Job) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Make a copy to avoid races on the data map
	cp := *job
	m.jobs[job.ID] = &cp
	return nil
}

func (m *MemoryStorage) GetJob(ctx context.Context, jobID string) (*models.Job, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	job, ok := m.jobs[jobID]
	if !ok {
		return nil, ErrJobNotFound
	}
	cp := *job
	return &cp, nil
}

func (m *MemoryStorage) ListReadyJobs(ctx context.Context, limit int) ([]*models.Job, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	now := time.Now()
	var ready []*models.Job
	for _, job := range m.jobs {
		if job.Status == "queued" && !job.NotBefore.After(now) {
			cp := *job
			ready = append(ready, &cp)
		}
	}

	// Sort by NotBefore ascending (oldest first)
	sort.Slice(ready, func(i, j int) bool {
		return ready[i].NotBefore.Before(ready[j].NotBefore)
	})

	if limit > 0 && len(ready) > limit {
		ready = ready[:limit]
	}
	return ready, nil
}

func (m *MemoryStorage) UpdateJobStatus(ctx context.Context, jobID string, status string, completedAt *time.Time, lastError string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	job, ok := m.jobs[jobID]
	if !ok {
		return ErrJobNotFound
	}
	job.Status = status
	job.LastError = lastError
	if completedAt != nil {
		t := *completedAt
		job.CompletedAt = &t
	}
	return nil
}

func (m *MemoryStorage) IncrementJobAttempts(ctx context.Context, jobID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	job, ok := m.jobs[jobID]
	if !ok {
		return ErrJobNotFound
	}
	job.Attempts++
	return nil
}

func (m *MemoryStorage) ClearJobs(ctx context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.jobs = make(map[string]*models.Job)
	return nil
}
