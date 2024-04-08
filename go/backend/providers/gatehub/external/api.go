package external

import "context"

type Client interface {
	IssueToken(ctx context.Context, userID string, product Product) (*IssueTokenResponse, error)
	CreateUser(ctx context.Context, email string) (*CreateUserResponse, error)
	GetOnboardingWidget(ctx context.Context, userID string) (string, error)
	GetOnOffRampWidget(ctx context.Context, userID string, isDeposit bool) (string, error)
	GetUserWallets(ctx context.Context, userID string) (*GetUserWalletsResponse, error)
}
