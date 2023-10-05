package ops

import (
	"context"
	"errors"

	"gitlab.com/fynbos/backend/payments"
	"gitlab.com/fynbos/backend/transactions"
)

func (a *Activity) SetReceiveTransactionID(ctx context.Context, paymentID, txID string) error {
	return setReceiveTransactionID(ctx, a.b, paymentID, txID)
}

func (a *Activity) UpdatePayInTransactionDestination(ctx context.Context, paymentID string) error {
	p, err := Lookup(ctx, a.b, paymentID)
	if err != nil {
		return err
	}

	if p.SendTransactionID == "" {
		return errors.New("no send transaction attached to payment")
	}

	w, err := lookupWallet(ctx, a.b, p.Receiver)
	if err != nil {
		return err
	}

	return a.b.Transactions().SetTransactionDestination(ctx, p.SendTransactionID, w.AddressString())
}

func (a *Activity) UpdatePayInTransactionState(ctx context.Context, paymentID string, state transactions.State) error {
	p, err := Lookup(ctx, a.b, paymentID)
	if err != nil {
		return err
	}

	if p.SendTransactionID == "" {
		return nil
	}

	return a.b.Transactions().SetTransactionState(ctx, p.SendTransactionID, state)
}

func (a *Activity) UpdatePayoutTransactionState(ctx context.Context, paymentID string, state transactions.State) error {
	p, err := Lookup(ctx, a.b, paymentID)
	if err != nil {
		return err
	}

	if p.ReceiveTransactionID == "" {
		return nil
	}

	return a.b.Transactions().SetTransactionState(ctx, p.ReceiveTransactionID, state)
}

func (a *Activity) AddPayInTransfer(ctx context.Context, paymentID, fkID string) error {
	p, err := Lookup(ctx, a.b, paymentID)
	if err != nil {
		return err
	}

	transferType := transactions.TransferTypeDebitCard
	switch p.Type {
	case payments.TypePeer2Peer:
		transferType = transactions.TransferTypeDebitCard
	case payments.TypeWebMonetization:
		transferType = transactions.TransferTypeDebitWebMonetization
	case payments.TypeReferral:
		transferType = transactions.TransferTypeDebitReferral
	}

	return a.b.Transactions().AddTransfers(ctx, p.SendTransactionID, []transactions.TransferArgs{
		{
			LinkedAccountID: p.SenderAccount,
			ForeignID:       fkID,
			Type:            transferType,
			Amount:          p.SenderAmount,
			State:           transactions.StateCompleted,
		},
	})
}

func (a *Activity) AddWebMonetizationPayInTransfer(ctx context.Context, paymentID string) error {
	p, err := Lookup(ctx, a.b, paymentID)
	if err != nil {
		return err
	}

	return a.b.Transactions().AddTransfers(ctx, p.SendTransactionID, []transactions.TransferArgs{
		{
			LinkedAccountID: p.SenderAccount,
			ForeignID:       paymentID,
			Type:            transactions.TransferTypeDebitWebMonetization,
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
		Note:                    p.Note,
		Source:                  senderWallet.AddressString(),
		Destination:             receiverWallet.AddressString(),
		Amount:                  p.ReceiverAmount,
		LinkedAccountTitle:      la.Title(),
		DestinationIdentity:     p.Receiver.Identifier,
		DestinationIdentityType: p.Receiver.Type.String(),
		Reference:               p.Note,
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

func (a *Activity) SetTransactionRefundState(ctx context.Context, paymentID string, state transactions.RefundState) error {
	p, err := Lookup(ctx, a.b, paymentID)
	if err != nil {
		return err
	}

	return a.b.Transactions().SetTransactionRefundState(ctx, p.SendTransactionID, state)
}
