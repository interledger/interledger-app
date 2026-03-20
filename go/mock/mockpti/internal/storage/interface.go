package storage

import (
	"context"
	"errors"

	"gitlab.com/fynbos/mock/mockpti/internal/models"
)

var (
	ErrUserNotFound       = errors.New("user not found")
	ErrAssessmentNotFound = errors.New("assessment not found")
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

	// Reset all data (for testing)
	Reset(ctx context.Context) error
}
