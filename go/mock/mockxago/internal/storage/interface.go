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

	// Reset all data (for testing)
	Reset(ctx context.Context) error
}
