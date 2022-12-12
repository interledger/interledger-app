package workflows

import (
	"context"

	"gitlab.com/fynbos/backend/transactions"
)

func (a *Activity) AddTransaction(ctx context.Context, args transactions.CreateTransactionArgs) error {
	return a.b.Transactions().CreateTransaction(ctx, args)
}

func (a *Activity) AddTransactionTransfer(ctx context.Context, args []transactions.CreateTransferArgs) error {
	return a.b.Transactions().AddTransfers(ctx, args)
}
