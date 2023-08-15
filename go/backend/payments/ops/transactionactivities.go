package ops

import (
	"context"

	"gitlab.com/fynbos/backend/transactions"
)

func (a *Activity) SetSendTransactionID(ctx context.Context, paymentID, txID string) error {
	return setSendTransactionID(ctx, a.b, paymentID, txID)
}

func (a *Activity) SetReceiveTransactionID(ctx context.Context, paymentID, txID string) error {
	return setReceiveTransactionID(ctx, a.b, paymentID, txID)
}

func (a *Activity) UpdateTransactionState(ctx context.Context, trxID string, state transactions.State) error {
	return a.b.Transactions().SetTransactionState(ctx, trxID, state)
}

func (a *Activity) AddPayInTransfer(ctx context.Context, paymentID, fkID string) error {
	p, err := Lookup(ctx, a.b, paymentID)
	if err != nil {
		return err
	}

	return a.b.Transactions().AddTransfers(ctx, p.SendTransactionID, []transactions.TransferArgs{
		{
			LinkedAccountID: p.SenderAccount,
			ForeignID:       fkID,
			Type:            transactions.TransferTypeDebitCard,
			Amount:          p.SenderAmount,
			State:           transactions.StateCompleted,
		},
	})
}

func (a *Activity) AddPayInRollbackTransfer(ctx context.Context, paymentID, fkID string) error {
	p, err := Lookup(ctx, a.b, paymentID)
	if err != nil {
		return err
	}

	return a.b.Transactions().AddTransfers(ctx, p.SendTransactionID, []transactions.TransferArgs{
		{
			LinkedAccountID: p.SenderAccount,
			ForeignID:       fkID,
			Type:            transactions.TransferTypeCreditCard,
			Amount:          p.SenderAmount,
			State:           transactions.StateCompleted,
		},
	})
}

func (a *Activity) CreatePayInTransaction(ctx context.Context, paymentID string) (string, error) {
	p, err := Lookup(ctx, a.b, paymentID)
	if err != nil {
		return "", err
	}

	// Already created, nothing to do
	if p.SendTransactionID != "" {
		return "", nil
	}

	receiverWallet, err := lookupWallet(ctx, a.b, p.Receiver)
	if err != nil {
		return "", err
	}

	senderWallet, err := lookupWallet(ctx, a.b, p.Sender)
	if err != nil {
		return "", err
	}

	la, err := a.b.LinkedAccounts().Get(ctx, p.SenderAccount)
	if err != nil {
		return "", err
	}

	return a.b.Transactions().CreateTransaction(ctx, transactions.CreateTransactionArgs{
		WalletID:                senderWallet.ID,
		ForeignID:               paymentID,
		ForeignType:             transactions.TransactionTypeOutgoing,
		Provider:                transactions.ProviderPaymentsEngine,
		State:                   transactions.StatePending,
		Note:                    "NOTE", // TODO : payment note
		Source:                  senderWallet.AddressString(),
		Destination:             receiverWallet.AddressString(),
		Amount:                  p.SenderAmount,
		LinkedAccountTitle:      la.Title(),
		DestinationIdentity:     p.Receiver.Identifier,
		DestinationIdentityType: p.Receiver.Type.String(),
		Reference:               "NOTE", // TODO : payment note
	})
}

func (a *Activity) CreatePayoutTransaction(ctx context.Context, paymentID, fkID string) (string, error) {
	p, err := Lookup(ctx, a.b, paymentID)
	if err != nil {
		return "", err
	}

	// Already created, nothing to do
	if p.ReceiveTransactionID != "" {
		return "", nil
	}

	receiverWallet, err := lookupWallet(ctx, a.b, p.Receiver)
	if err != nil {
		return "", err
	}

	senderWallet, err := lookupWallet(ctx, a.b, p.Sender)
	if err != nil {
		return "", err
	}

	la, err := a.b.LinkedAccounts().Get(ctx, p.ReceiverAccount)
	if err != nil {
		return "", err
	}

	return a.b.Transactions().CreateTransaction(ctx, transactions.CreateTransactionArgs{
		WalletID:                receiverWallet.ID,
		ForeignID:               paymentID,
		ForeignType:             transactions.TransactionTypeIncoming,
		Provider:                transactions.ProviderPaymentsEngine,
		State:                   transactions.StatePending,
		Note:                    "NOTE", // TODO : payment note
		Source:                  senderWallet.AddressString(),
		Destination:             receiverWallet.AddressString(),
		Amount:                  p.ReceiverAmount,
		LinkedAccountTitle:      la.Title(),
		DestinationIdentity:     p.Receiver.Identifier,
		DestinationIdentityType: p.Receiver.Type.String(),
		Reference:               "NOTE", // TODO : payment note
		Transfers: []transactions.TransferArgs{
			{
				LinkedAccountID: la.ID,
				ForeignID:       fkID,
				Type:            transactions.TransferTypeCreditCard, // TODO: infer these from the send/receive account types
				Amount:          p.ReceiverAmount,
				State:           transactions.StateCompleted,
			},
		},
	})
}
