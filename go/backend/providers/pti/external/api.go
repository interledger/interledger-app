package external

import (
	"context"
	"encoding/json"
)

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
	WalletWithdrawal(ctx context.Context, args WithdrawalArgs) (*WithdrawDetails, error)
	UpdateTransactionStatus(ctx context.Context, args UpdateTxStatusArgs) (string, error)
	StartTransferAssessment(ctx context.Context, args TransferArgs) (*IDResponse, error)
	GetTransactionAssessment(ctx context.Context, requestID string) (*TransactionAssessment, error)
	CreateTransfer(ctx context.Context, args TransferArgs) (*IDResponse, error)
	GetTransaction(ctx context.Context, requestID string) (*TransactionStatus, error)
	CreateJWT(ctx context.Context, args TokenArgs) (*TokenResponse, error)
	GetUsersPaymentInformation(ctx context.Context, userID, id string) (json.RawMessage, error)
	CreateBankAccount(ctx context.Context, userID string, args BankAccountPaymentInformation) (*BankAccountPaymentInformation, error)
	// CreateBankAccountFromPlaid registers a bank account on Fiant using a
	// Plaid processor token instead of raw ACH fields. Fiant calls Plaid
	// server-side to resolve the underlying account details
	CreateBankAccountFromPlaid(ctx context.Context, userID, processorToken string) (*BankAccountPaymentInformation, error)
}
