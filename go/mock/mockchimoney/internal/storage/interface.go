package storage

import (
	"context"
	"errors"

	"gitlab.com/fynbos/mock/mockchimoney/internal/models"
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
}
