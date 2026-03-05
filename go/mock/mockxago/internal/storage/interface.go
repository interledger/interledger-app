package storage

import (
	"context"
	"errors"
	"time"

	"gitlab.com/fynbos/mock/mockxago/internal/models"
)

var (
	ErrTokenNotFound        = errors.New("token not found")
	ErrInvalidToken         = errors.New("invalid token")
	ErrTokenExpired         = errors.New("token expired")
	ErrSubAccountNotFound   = errors.New("sub-account not found")
	ErrBeneficiaryNotFound  = errors.New("beneficiary not found")
	ErrInsufficientBalance  = errors.New("insufficient balance")
	ErrDuplicateTransaction = errors.New("duplicate transaction")
)

// Storage interface defines all storage operations
type Storage interface {
	// Token operations
	SaveAccessToken(ctx context.Context, token *models.AccessToken) error
	GetAccessToken(ctx context.Context, tokenValue string) (*models.AccessToken, error)
	InvalidateAccessToken(ctx context.Context, tokenValue string) error
	SaveTokenAccount(ctx context.Context, tokenValue string, accountID string) error
	GetAccountIDByToken(ctx context.Context, tokenValue string) (string, error)

	// Sub-account operations
	SaveSubAccount(ctx context.Context, account *models.SubAccount) error
	GetSubAccount(ctx context.Context, accountID string) (*models.SubAccount, error)
	GetSubAccountByWalletID(ctx context.Context, walletID string) (*models.SubAccount, error)
	UpdateSubAccount(ctx context.Context, account *models.SubAccount) error

	// Beneficiary operations
	SaveBeneficiary(ctx context.Context, beneficiary *models.Beneficiary) error
	GetBeneficiary(ctx context.Context, beneficiaryID string) (*models.Beneficiary, error)
	ListBeneficiariesByWallet(ctx context.Context, walletID string, limit int, offset int) ([]*models.Beneficiary, int, error)
	UpdateBeneficiaryStatus(ctx context.Context, beneficiaryID string, status string) error

	// Transaction operations
	SaveTransaction(ctx context.Context, transaction *models.Transaction) error
	GetTransaction(ctx context.Context, transactionID string) (*models.Transaction, error)
	GetTransactionByIdempotencyKey(ctx context.Context, key string) (*models.Transaction, error)
	SaveIdempotencyKey(ctx context.Context, key string, transactionID string) error
	ListTransactionsByAccount(ctx context.Context, accountID string, limit int, offset int) ([]*models.Transaction, int, error)
	UpdateTransactionStatus(ctx context.Context, transactionID string, status string) error

	// Balance operations
	GetBalance(ctx context.Context, walletID string, currency string) (available float64, reserved float64, err error)
	SetBalance(ctx context.Context, walletID string, currency string, available float64, reserved float64) error
	AddBalance(ctx context.Context, walletID string, currency string, amount float64) error
	SubtractBalance(ctx context.Context, walletID string, currency string, amount float64) error

	// Deposit operations
	SaveDeposit(ctx context.Context, deposit *models.Deposit) error
	GetDeposit(ctx context.Context, depositID string) (*models.Deposit, error)
	GetDepositByReference(ctx context.Context, reference string) (*models.Deposit, error)
	ListDeposits(ctx context.Context, limit int, offset int) ([]*models.Deposit, int, error)
	UpdateDepositStatus(ctx context.Context, depositID string, status string) error
	ClearDeposits(ctx context.Context) error
	ClearTransactions(ctx context.Context) error
	ClearBalances(ctx context.Context) error

	// Job operations
	SaveJob(ctx context.Context, job *models.Job) error
	GetJob(ctx context.Context, jobID string) (*models.Job, error)
	ListReadyJobs(ctx context.Context, limit int) ([]*models.Job, error)
	UpdateJobStatus(ctx context.Context, jobID string, status string, completedAt *time.Time, lastError string) error
	IncrementJobAttempts(ctx context.Context, jobID string) error
	ClearJobs(ctx context.Context) error

	// Reset all data (for testing)
	Reset(ctx context.Context) error
}
