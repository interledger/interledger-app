package xago

import (
	"context"
	"net/http"

	"gitlab.com/fynbos/backend/currency"
)

type Client interface {
	WebhookHandler() http.HandlerFunc
	CreateBeneficiary(ctx context.Context, bankAcc CreateBankAccountArgs) (Await, error)
	CreateBalanceAccount(ctx context.Context, args CreateBalanceAccArgs) (Await, error)
	CreateTransaction(ctx context.Context, args CreateTransactionArgs) (*Transaction, error)
	GetBalance(ctx context.Context, linkedAccountID string) (*Balance, error)
	ReserveBalance(ctx context.Context, linkedAccountID, txID string, amt currency.Amount) (*Balance, error)
	FinaliseReserve(ctx context.Context, txID string) error
	RollbackReserve(ctx context.Context, txID string) error
	AssignBalance(ctx context.Context, linkedAccountID, txID string, amt currency.Amount) (*Balance, error)
}
