package user

import (
	"context"
	"time"
)

type Client interface {
	UserForCookie(ctx context.Context, cookie string) (*User, error)
	UserForToken(ctx context.Context, token string) (*User, error)
	UserForContext(ctx context.Context) (*User, error)
	GetUser(ctx context.Context, userID string) (*User, error)
	ListUsers(ctx context.Context, walletID string) ([]User, error)
	CheckUserTotpEnabled(ctx context.Context, identityID string) (bool, error)
	Delete2FATotpEnrollment(ctx context.Context, identityID string) error
	ResetEmailPassword(ctx context.Context, identityID string) (string, error)
	GetTotpURL(ctx context.Context, userID string) (string, error)
	ValidateTotpCode(ctx context.Context, userID, code string, now time.Time) error
	GetUserIDForWallet(ctx context.Context, walletID string) (string, error)
	SetPhoneVerified(ctx context.Context, userID string) error
	UpdateUserPhone(ctx context.Context, userID string, phone string) error
	// FindWalletIDByEmail resolves a Kratos credential identifier (email) to a
	// wallet ID via the user_wallets table. Returns "" if no match is found.
	FindWalletIDByEmail(ctx context.Context, email string) (string, error)
	// FindWalletIDsByIdentifierPrefix resolves every Kratos identity whose
	// credential identifier (email or phone) starts with term to its wallet
	// ID(s). Returns nil for zero matches.
	FindWalletIDsByIdentifierPrefix(ctx context.Context, term string) ([]string, error)
}
