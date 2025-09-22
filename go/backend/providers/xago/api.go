package xago

import (
	"context"
	"net/http"
	"time"

	"gitlab.com/fynbos/backend/currency"
)

type Client interface {
	WebhookHandler() http.HandlerFunc
	LookupSubAccount(ctx context.Context, walletID string) (*SubAccount, error)
	CreateBeneficiary(ctx context.Context, bankAcc CreateBankAccountArgs) (Await, error)
	CreateBalanceAccount(ctx context.Context, args CreateBalanceAccArgs) (Await, error)
	CreateTransaction(ctx context.Context, args CreateTransactionArgs) (*Transaction, error)
	LookupWithdrawal(ctx context.Context, id string) (*Withdrawal, error)
	GetBalance(ctx context.Context, linkedAccountID string) (*Balance, error)
	ReserveBalance(ctx context.Context, linkedAccountID, txID string, amt currency.Amount, timeout time.Duration) (*Balance, error)
	FinaliseReserve(ctx context.Context, txID string) error
	RollbackReserve(ctx context.Context, txID string) error
	AssignBalance(ctx context.Context, linkedAccountID, txID string, amt currency.Amount) (*Balance, error)
	TestDeposit(ctx context.Context, sa SubAccount) error
	UpdateInquiryLink(ctx context.Context, accountID string, walletID string) error
}
