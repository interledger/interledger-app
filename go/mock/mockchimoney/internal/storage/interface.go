package storage

import (
	"context"
	"errors"

	"github.com/interledger/interledger-app/go/mock/mockchimoney/internal/models"
)

var (
	// ErrNotFound indicates a requested entity is not in storage.
	ErrNotFound = errors.New("not found")
	// ErrAlreadyExists indicates an insert attempted to use an existing key.
	ErrAlreadyExists = errors.New("already exists")
)

// Store defines persistence operations for MockChimoney resources.
type Store interface {
	CreateSubAccount(ctx context.Context, account models.SubAccount) (models.SubAccount, error)
	GetSubAccount(ctx context.Context, id string) (models.SubAccount, error)
	ListSubAccounts(ctx context.Context) ([]models.SubAccount, error)
	UpdateSubAccountKYCStatus(ctx context.Context, id string, status string) (models.SubAccount, error)

	CreatePayment(ctx context.Context, payment models.Payment) (models.Payment, error)
	GetPaymentByIssueID(ctx context.Context, issueID string) (models.Payment, error)
	UpdatePaymentStatus(ctx context.Context, issueID string, status string) (models.Payment, error)

	CreatePayout(ctx context.Context, payout models.Payout) (models.Payout, error)
	GetPayoutByChiRef(ctx context.Context, chiRef string) (models.Payout, error)
	GetPayoutByIssueID(ctx context.Context, issueID string) (models.Payout, error)
	UpdatePayoutStatus(ctx context.Context, issueID string, status string) (models.Payout, error)
}
