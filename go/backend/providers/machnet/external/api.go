package external

import "context"

type Client interface {
	RegisterUser(ctx context.Context, user User) (*User, error)
	UpdateUser(ctx context.Context, id string, newValues User) (*User, error)
	GetUserByID(ctx context.Context, id string) (*User, error)
	InitiateKyc(ctx context.Context, userID string) error
	GetVerificationStatus(ctx context.Context, userID string) (*VerificationStatus, error)
	GetReceiveUserList(ctx context.Context, userID string) ([]User, error)
	GetFundingAccountWidgetToken(ctx context.Context, userID string) (*WidgetTokenResponse, error)
	GetUserFundingsource(ctx context.Context, userID, fundingsourceID string) (*FundingSource, error)
	CreateTransaction(ctx context.Context, transaction Transaction) (*Transaction, error)
}
