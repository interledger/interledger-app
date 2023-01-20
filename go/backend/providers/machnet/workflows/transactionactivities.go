package workflows

import (
	"context"

	"gitlab.com/fynbos/backend/transactions"
)

func (a *Activity) GetTransaction(ctx context.Context, walletID string, trxID string) (transactions.Transaction, error) {
	trx, err := a.b.Transactions().GetTransaction(ctx, walletID, trxID)
	if err != nil {
		return transactions.Transaction{}, err
	}

	return *trx, nil
}

func (a *Activity) AddTransaction(ctx context.Context, args transactions.CreateTransactionArgs) (string, error) {
	return a.b.Transactions().CreateTransaction(ctx, args)
}

func (a *Activity) AddTransactionTransfer(ctx context.Context, trxID string, args []transactions.TransferArgs) error {
	return a.b.Transactions().AddTransfers(ctx, trxID, args)
}

func (a *Activity) UpdateTransactionState(ctx context.Context, trxID string, state transactions.State) error {
	return a.b.Transactions().SetTransactionState(ctx, trxID, state)
}

func (a *Activity) UpdateTransferState(ctx context.Context, tfrID string, state transactions.State) error {
	return a.b.Transactions().SetTransferState(ctx, tfrID, state)
}

func (a *Activity) UpdateTransferStateByType(ctx context.Context, trxID string, walletID string, tfsType transactions.TransferType, state transactions.State) error {
	trx, err := a.b.Transactions().GetTransaction(ctx, walletID, trxID)
	if err != nil {
		return err
	}
	for _, tfr := range trx.Transfers {
		if tfr.Type == tfsType {
			err = a.b.Transactions().SetTransferState(ctx, tfr.ID, state)
			if err != nil {
				return err
			}
		}
	}

	return nil
}
