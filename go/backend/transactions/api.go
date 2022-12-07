package transactions

import (
	"context"
)

type Client interface {
	CreateTransaction(ctx context.Context, args CreateTransactionArgs) error
	AddTransfer(ctx context.Context, args CreateTransferArgs) error
}
