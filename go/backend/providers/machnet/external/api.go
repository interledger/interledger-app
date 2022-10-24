package external

import "context"

type Client interface {
	RegisterUser(ctx context.Context, user User) (*User, error)
	UpdateUser(ctx context.Context, id string, newValues User) (*User, error)
	GetUserByID(ctx context.Context, id string) (*User, error)
	InitiateKYC(ctx context.Context, userID string) (*InitiateKycResponse, error)
	GetVerificationStatus(ctx context.Context, userID string) (*VerificationStatus, error)
	GetReceiveUserList(ctx context.Context, userID string) ([]User, error)
	GetFundingAccountWidgetToken(ctx context.Context, userID string) (*WidgetTokenResponse, error)
	GetUserFundingsource(ctx context.Context, userID, fundingsourceID string) (*FundingSource, error)
	CreateTransaction(ctx context.Context, transaction CreateTransactionArgs) (*Transaction, error)
	GetUserTransaction(ctx context.Context, userID, id string) (*Transaction, error)
	UpdateDeliveryRequest(ctx context.Context, request DeliveryRequest) error
	CreateReceiveUserBankAccount(ctx context.Context, sendUserID, receiveUserID string, acc ReceiveUserBankAccount) (*ReceiveUserBankAccount, error)
	GetBanks(ctx context.Context, countryCode string) ([]Bank, error)
	ListReceiveUserBankAccounts(ctx context.Context, sendUserID, receiveUserID string) ([]ReceiveUserBankAccount, error)
}
