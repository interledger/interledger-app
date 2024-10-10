package external

import "context"

type Client interface {
	CreateUser(ctx context.Context, args CreateUserArgs) (string, error)
	GetUser(ctx context.Context, id string) (*User, error)
	PatchUser(ctx context.Context, args PatchUserArgs) (string, error)
	PutUser(ctx context.Context, args PutUserArgs) (string, error)
	CreateWallet(ctx context.Context, args CreateWalletArgs) (*Wallet, error)
	GetWallet(ctx context.Context, userID, id string) (*Wallet, error)
	ListWallets(ctx context.Context, userID string) ([]Wallet, error)
	StartUserAssessment(ctx context.Context, args StartUserAssessmentArgs) (string, error)
	GetUserAssessment(ctx context.Context, userID string) (*Assessment, error)
	WalletDeposit(ctx context.Context, args DepositArgs) (string, error)
	WalletWithdrawal(ctx context.Context, args WithdrawalArgs) (string, error)
	UpdateTransactionStatus(ctx context.Context, args UpdateTxStatusArgs) (string, error)
	StartTransferAssessment(ctx context.Context, args TransferArgs) (*IDResponse, error)
	GetTransactionAssessment(ctx context.Context, requestID string) (*TransactionAssessment, error)
	CreateTransfer(ctx context.Context, args TransferArgs) (*IDResponse, error)
	GetTransaction(ctx context.Context, requestID string) (*TransactionStatus, error)
	CreateJWT(ctx context.Context, args TokenArgs) (*TokenResponse, error)
}
