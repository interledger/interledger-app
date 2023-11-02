package xago

import (
	"context"
	"net/http"
)

type Client interface {
	WebhookHandler() http.HandlerFunc
	CreateBeneficiary(ctx context.Context, bankAcc CreateBankAccountArgs) (Await, error)
	CreateTransaction(ctx context.Context, args CreateTransactionArgs) (*Transaction, error)
}
