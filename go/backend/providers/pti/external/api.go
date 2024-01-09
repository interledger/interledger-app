package external

import "context"

type Client interface {
	CreateUser(ctx context.Context, args CreateUserArgs) (string, error)
	CreateWallet(ctx context.Context, args CreateWalletArgs) (*Wallet, error)
	GetWallet(ctx context.Context, userID, id string) (*Wallet, error)
	StartUserAssessment(ctx context.Context, args StartUserAssessmentArgs) (string, error)
}
