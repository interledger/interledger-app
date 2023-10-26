package user

import (
	"context"
)

type Client interface {
	UserForToken(ctx context.Context, token string) (*User, error)
	UserForCookie(ctx context.Context, cookie string) (*User, error)
	UserForContext(ctx context.Context) (*User, error)
	GetUser(ctx context.Context, userID string) (*User, error)
	ListUsers(ctx context.Context, walletID string) ([]User, error)
}
