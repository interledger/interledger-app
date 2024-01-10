package external

import "context"

type Client interface {
	CreateUser(ctx context.Context, args CreateUserArgs) (string, error)
	GetUser(ctx context.Context, id string) (*User, error)
	CreateWallet(ctx context.Context, args CreateWalletArgs) (*Wallet, error)
	GetWallet(ctx context.Context, userID, id string) (*Wallet, error)
	StartUserAssessment(ctx context.Context, args StartUserAssessmentArgs) (string, error)
	GetUserAssessment(ctx context.Context, userID string) (*Assessment, error)
}
