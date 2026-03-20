package storage

import (
	"context"
	"errors"

	"gitlab.com/fynbos/mock/mockpti/internal/models"
)

var (
	ErrUserNotFound               = errors.New("user not found")
	ErrAssessmentNotFound         = errors.New("assessment not found")
	ErrWalletNotFound             = errors.New("wallet not found")
	ErrPaymentInformationNotFound = errors.New("payment information not found")
	ErrTransactionNotFound        = errors.New("transaction not found")
)

// Storage defines all persistence operations for mockpti.
type Storage interface {
	// User operations
	SaveUser(ctx context.Context, user *models.User) error
	GetUser(ctx context.Context, userID string) (*models.User, error)
	UpdateUser(ctx context.Context, user *models.User) error

	// Assessment operations
	SaveAssessment(ctx context.Context, assessment *models.Assessment) error
	GetLatestAssessment(ctx context.Context, userID string) (*models.Assessment, error)

	// Wallet operations
	SaveWallet(ctx context.Context, wallet *models.Wallet) error
	GetWallet(ctx context.Context, userID, walletID string) (*models.Wallet, error)
	ListWallets(ctx context.Context, userID string) ([]*models.Wallet, error)

	// Payment information operations
	SavePaymentInformation(ctx context.Context, pi *models.PaymentInformation) error
	GetPaymentInformation(ctx context.Context, userID, piID string) (*models.PaymentInformation, error)

	// Transaction operations
	SaveTransaction(ctx context.Context, tx *models.Transaction) error
	GetTransaction(ctx context.Context, requestID string) (*models.Transaction, error)
	SaveTransactionUpdate(ctx context.Context, update *models.TransactionUpdate) error

	// Reset all data (for testing)
	Reset(ctx context.Context) error
}
