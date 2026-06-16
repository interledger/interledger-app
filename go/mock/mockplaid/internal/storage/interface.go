package storage

import (
	"context"
	"errors"

	"gitlab.com/fynbos/mock/mockplaid/internal/models"
)

// Storage errors.
var (
	ErrLinkSessionNotFound = errors.New("link session not found")
	ErrItemNotFound        = errors.New("item not found")
)

// Storage is the mockplaid state seam: link sessions, items (resolvable by
// public token or access token), and a monotonic account-id sequence used to
// mint always-new account IDs for the "fresh" mock bank.
type Storage interface {
	SaveLinkSession(ctx context.Context, s models.LinkSession) error
	GetLinkSession(ctx context.Context, linkToken string) (models.LinkSession, error)

	SaveItem(ctx context.Context, item models.Item) error
	GetItemByPublicToken(ctx context.Context, publicToken string) (models.Item, error)
	GetItemByAccessToken(ctx context.Context, accessToken string) (models.Item, error)
	DeleteItemByAccessToken(ctx context.Context, accessToken string) error

	// NextAccountSeq returns a strictly increasing counter value, used to mint
	// unique account IDs for the always-new mock bank.
	NextAccountSeq(ctx context.Context) (uint64, error)

	// Reset wipes all state (test convenience).
	Reset(ctx context.Context) error
}
