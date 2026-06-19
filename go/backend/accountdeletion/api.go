package accountdeletion

import "context"

type Client interface {
	// Returns ErrAlreadyRequested if a request already exists for this user.
	Request(ctx context.Context, userID string) error
	// Returns nil, nil when no request exists.
	GetForUser(ctx context.Context, userID string) (*Request, error)
	// No-op when no request exists.
	Delete(ctx context.Context, userID string) error
}
