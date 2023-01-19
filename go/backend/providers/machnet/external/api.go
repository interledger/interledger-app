package external

import "context"

type Client interface {
	RegisterUser(ctx context.Context, user User) (*User, error)
	UpdateUser(ctx context.Context, id string, newValues User) (*User, error)
	GetUserByID(ctx context.Context, id string) (*User, error)
	GetUserLimits(ctx context.Context, id string) ([]UserLimit, error)
	InitiateKYC(ctx context.Context, userID string) (*InitiateKycResponse, error)
	GetVerificationStatus(ctx context.Context, userID string) (*VerificationStatus, error)
	GetReceiveUserList(ctx context.Context, userID string) ([]User, error)
	GetFundingAccountWidgetToken(ctx context.Context, userID string) (*WidgetTokenResponse, error)
	GetUserFundingsource(ctx context.Context, userID, fundingsourceID string) (*FundingSource, error)
	DeleteFundingSource(ctx context.Context, userID, fundingSourceID string) error
	CreateTransaction(ctx context.Context, transaction CreateTransactionArgs) (*Transaction, error)
	GetUserTransaction(ctx context.Context, userID, id string) (*Transaction, error)
	UpdateDeliveryRequest(ctx context.Context, request DeliveryRequest) error
	CreateReceiveUserBankAccount(ctx context.Context, sendUserID, receiveUserID string, acc ReceiveUserBankAccount) (*ReceiveUserBankAccount, error)
	GetBanks(ctx context.Context, countryCode string) ([]Bank, error)
	ListReceiveUserBankAccounts(ctx context.Context, sendUserID, receiveUserID string) ([]ReceiveUserBankAccount, error)
	CreateUserWallet(ctx context.Context, sendUserID, nickName string) (*Wallet, error)
	GetUserWallet(ctx context.Context, sendUserID, walletID string) (*Wallet, error)
	FundUserWallet(ctx context.Context, args FundWalletArgs) (*FundWalletResponse, error)
	CreateWalletTransfer(ctx context.Context, args WalletTransferArgs) (*WalletTransfer, error)
	WithdrawFromUserWallet(ctx context.Context, args WithdrawFromUserWalletArgs) (*WalletWithdrawal, error)
}
