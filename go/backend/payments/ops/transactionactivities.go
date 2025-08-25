package ops

import (
	"context"
	"errors"

	"gitlab.com/fynbos/backend/providers/pti"

	"gitlab.com/fynbos/backend/providers/astra"

	"gitlab.com/fynbos/backend/providers/xago"

	"gitlab.com/fynbos/backend/currency"
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
	case payments.TypeWithdrawal:
		transferType = transactions.TransferTypeDebitBalance
	case payments.TypeDeposit:
		transferType = transactions.TransferTypeDebitCard
	case payments.TypePeer2Peer:
		transferType = transactions.TransferTypeDebitBalance
	case payments.TypeWebMonetization:
		transferType = transactions.TransferTypeDebitWebMonetization
	}

	return a.b.Transactions().AddTransfers(ctx, p.SendTransactionID, []transactions.TransferArgs{
		{
			LinkedAccountID: p.SenderAccount,
			ForeignID:       fkID,
			Type:            transferType,
			Amount:          p.SenderAmount,
			ProviderFee:     currency.FromFloat64(0, currency.USD),
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
			ProviderFee:     currency.FromFloat64(0, currency.USD),
			State:           transactions.StateCompleted,
		},
	})
}

func (a *Activity) AddWithdrawalTransfer(ctx context.Context, paymentID string) error {
	p, err := Lookup(ctx, a.b, paymentID)
	if err != nil {
		return err
	}

	if p.Type != payments.TypeWithdrawal {
		return nil
	}

	la, err := a.b.LinkedAccounts().Get(ctx, p.ReceiverAccount)
	if err != nil {
		return err
	}

	typ := transactions.TransferTypeCreditBankAccount
	if la.Provider == astra.ProviderName {
		typ = transactions.TransferTypeCreditCard
	}

	return a.b.Transactions().AddTransfers(ctx, p.SendTransactionID, []transactions.TransferArgs{
		{
			LinkedAccountID: p.ReceiverAccount,
			ForeignID:       paymentID,
			Type:            typ,
			Amount:          p.ReceiverAmount,
			ProviderFee:     currency.FromFloat64(0, currency.USD),
			State:           transactions.StateCompleted,
		},
	})
}

func (a *Activity) AddDepositTransfer(ctx context.Context, paymentID string) error {
	p, err := Lookup(ctx, a.b, paymentID)
	if err != nil {
		return err
	}

	if p.Type != payments.TypeDeposit {
		return nil
	}

	la, err := a.b.LinkedAccounts().Get(ctx, p.ReceiverAccount)
	if err != nil {
		return err
	}

	if la.Provider == astra.ProviderName {
		return nil
	}

	return a.b.Transactions().AddTransfers(ctx, p.SendTransactionID, []transactions.TransferArgs{
		{
			LinkedAccountID: p.ReceiverAccount,
			ForeignID:       paymentID,
			Type:            transactions.TransferTypeCreditBalance,
			Amount:          p.ReceiverAmount,
			ProviderFee:     currency.FromFloat64(0, currency.USD),
			State:           transactions.StateCompleted,
		},
	})
}

func (a *Activity) AddPayInRollbackTransfer(ctx context.Context, paymentID, fkID string) error {
	p, err := Lookup(ctx, a.b, paymentID)
	if err != nil {
		return err
	}

	typ := transactions.TransferTypeCreditCard
	if p.Type == payments.TypeWithdrawal {
		typ = transactions.TransferTypeCreditBalance
	} else {
		la, err := a.b.LinkedAccounts().Get(ctx, p.SenderAccount)
		if err != nil {
			return err
		}
		if la.Provider == xago.ProviderName || la.Provider == pti.ProviderName && la.Type == pti.AccTypeBalance {
			typ = transactions.TransferTypeCreditBalance
		}
	}

	return a.b.Transactions().AddTransfers(ctx, p.SendTransactionID, []transactions.TransferArgs{
		{
			LinkedAccountID: p.SenderAccount,
			ForeignID:       fkID,
			Type:            typ,
			Amount:          p.SenderAmount,
			ProviderFee:     currency.FromFloat64(0, currency.USD),
			State:           transactions.StateCompleted,
		},
	})
}

func (a *Activity) CreatePayoutTransaction(ctx context.Context, paymentID, txID, fkID string) (string, error) {
	p, err := Lookup(ctx, a.b, paymentID)
	if err != nil {
		return "", err
	}

	// Already created and we do not create payout transactions for withdrawals, nothing to do
	if p.ReceiveTransactionID != "" || p.Type == payments.TypeWithdrawal || p.Type == payments.TypeDeposit {
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

	transType := transactions.TransferTypeCreditCard
	if la.Provider == xago.ProviderName || la.Provider == pti.ProviderName {
		transType = transactions.TransferTypeCreditBalance
	}

	txType := transactions.TransactionTypeReceived
	var title string
	if p.Type == payments.TypeWebMonetization {
		txType = transactions.TransactionTypeWebMonetizationIncoming
		if senderWallet != nil {
			title = senderWallet.Name
		}
	}

	return a.b.Transactions().CreateTransaction(ctx, transactions.CreateTransactionArgs{
		ID:                      txID,
		WalletID:                receiverWallet.ID,
		ForeignID:               paymentID,
		ForeignType:             txType,
		Provider:                transactions.ProviderPaymentsEngine,
		State:                   transactions.StatePending,
		Note:                    p.Note,
		Source:                  senderWallet.AddressString(),
		Destination:             receiverWallet.AddressString(),
		Amount:                  p.ReceiverAmount,
		ProviderFee:             currency.FromFloat64(0, currency.USD),
		LinkedAccountTitle:      la.Title(),
		DestinationIdentity:     p.Receiver.Identifier,
		DestinationIdentityType: p.Receiver.Type.String(),
		Reference:               p.Note,
		Transfers: []transactions.TransferArgs{
			{
				LinkedAccountID: la.ID,
				ForeignID:       fkID,
				Type:            transType,
				Amount:          p.ReceiverAmount,
				ProviderFee:     currency.FromFloat64(0, currency.USD),
				State:           transactions.StateCompleted,
			},
		},
		Title: title,
	})
}

func (a *Activity) SetTransactionRefundState(ctx context.Context, paymentID string, state transactions.RefundState) error {
	p, err := Lookup(ctx, a.b, paymentID)
	if err != nil {
		return err
	}

	if p.Type == payments.TypeWithdrawal {
		return nil
	}

	return a.b.Transactions().SetTransactionRefundState(ctx, p.SendTransactionID, state)
}
