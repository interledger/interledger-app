package workflows

import (
	"context"

	"gitlab.com/fynbos/backend/transactions"
)

func (a *Activity) AddTransaction(ctx context.Context, args transactions.CreateTransactionArgs) error {
	return a.b.Transactions().CreateTransaction(ctx, args)
}

func (a *Activity) AddTransactionTransfer(ctx context.Context, args []transactions.TransferArgs) error {
	return a.b.Transactions().AddTransfers(ctx, args)
}

func (a *Activity) UpdateTransactionTransfer(ctx context.Context, args []transactions.TransferArgs) error {
	return a.b.Transactions().UpdateTransfers(ctx, args)
}

func (a *Activity) UpdateTransactionState(ctx context.Context, args transactions.UpdateTransactionArgs) error {
	return a.b.Transactions().UpdateTransaction(ctx, args)
}
