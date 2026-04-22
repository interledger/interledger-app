package user

import (
	"context"
)

type Client interface {
	UserForCookie(ctx context.Context, cookie string) (*User, error)
	UserForToken(ctx context.Context, token string) (*User, error)
	UserForContext(ctx context.Context) (*User, error)
	GetUser(ctx context.Context, userID string) (*User, error)
	ListUsers(ctx context.Context, walletID string) ([]User, error)
	CheckUserTotpEnabled(ctx context.Context, identityID string) (bool, error)
	Delete2FATotpEnrollment(ctx context.Context, identityID string) error
	GetTotpURL(ctx context.Context, userID string) (string, error)
	GetUserIDForWallet(ctx context.Context, walletID string) (string, error)
	SetPhoneVerified(ctx context.Context, userID string) error
}
