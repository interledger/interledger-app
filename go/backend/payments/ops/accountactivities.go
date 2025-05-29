package ops

import (
	"context"

	"github.com/google/uuid"

	"gitlab.com/fynbos/backend/providers/xago"

	"gitlab.com/fynbos/backend/payments"
)

func (a *Activity) WithdrawFromXagoBalance(ctx context.Context, paymentID string) (string, error) {
	p, err := Lookup(ctx, a.b, paymentID)
	if err != nil {
		return "", err
	}

	if p.Type != payments.TypeWithdrawal {
		return uuid.NewString(), nil
	}

	tx, err := a.b.Xago().CreateTransaction(ctx, xago.CreateTransactionArgs{
		WalletID:        p.Sender.WalletID,
		LinkedAccountID: p.ReceiverAccount,
		TransactionID:   p.SendTransactionID,
		Amount:          p.ReceiverAmount,
		Reference:       "Interledger - " + p.PublicID,
	})
	if err != nil {
		return "", err
	}

	return tx.ID, nil

}
