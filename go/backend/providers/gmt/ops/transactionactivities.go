package ops

import (
	"context"

	"gitlab.com/fynbos/backend/providers"

	"gitlab.com/fynbos/backend/transactions"
)

func (a *Activity) AddIncomingTransaction(ctx context.Context, pArgs providers.TransfersArgs, args transactions.CreateTransactionArgs) (string, error) {
	outgoing, err := a.b.Transactions().GetTransaction(ctx, pArgs.FromWalletID, pArgs.FromTransactionID)
	if err != nil {
		return "", err
	}

	la, err := a.b.LinkedAccounts().Get(ctx, pArgs.ToLinkedAccountID)
	if err != nil {
		return "", err
	}

	args.DestinationIdentityType = outgoing.DestinationIdentityType
	args.DestinationIdentity = outgoing.Destination
	args.LinkedAccountTitle = la.Title()
	args.Reference = outgoing.Reference

	return a.b.Transactions().CreateTransaction(ctx, args)
}

func (a *Activity) AddTransactionTransfer(ctx context.Context, trxID string, args []transactions.TransferArgs) error {
	return a.b.Transactions().AddTransfers(ctx, trxID, args)
}

func (a *Activity) UpdateTransactionState(ctx context.Context, trxID string, state transactions.State) error {
	return a.b.Transactions().SetTransactionState(ctx, trxID, state)
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
