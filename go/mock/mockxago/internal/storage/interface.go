package storage

import (
	"context"
	"errors"

	"gitlab.com/fynbos/mock/mockxago/internal/models"
)

var (
	ErrTokenNotFound       = errors.New("token not found")
	ErrInvalidToken        = errors.New("invalid token")
	ErrTokenExpired        = errors.New("token expired")
	ErrSubAccountNotFound  = errors.New("sub-account not found")
	ErrBeneficiaryNotFound = errors.New("beneficiary not found")
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
	ListBeneficiariesByWallet(ctx context.Context, accountID string, limit int, offset int) ([]*models.Beneficiary, int, error)
	UpdateBeneficiaryStatus(ctx context.Context, beneficiaryID string, status string) error

	// Reset all data (for testing)
	Reset(ctx context.Context) error
}
