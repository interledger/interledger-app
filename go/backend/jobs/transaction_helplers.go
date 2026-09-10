package jobs

import (
	"context"

	"github.com/interledger/interledger-app/go/backend/payments"
	"github.com/interledger/interledger-app/go/backend/transactions"
)

func (a *Activity) UpdateTransactionStateJob(ctx context.Context, transactionID string, state transactions.State) error {
	return a.b.Transactions().SetTransactionState(ctx, transactionID, state)

}

func (a *Activity) UpdatePaymentStateJob(ctx context.Context, paymentID string, state payments.State) error {
	_, err := a.b.DB().ExecContext(ctx, "UPDATE payments SET state=$1, updated_at=now() where id=$2", state, paymentID)
	if err != nil {
		return err
	}
	return nil
}

func (a *Activity) GetTransactionByID(ctx context.Context, data CheckTransaction) (*transactions.Transaction, error) {
	return a.b.Transactions().GetTransaction(ctx, data.WalletID, data.TransactionID)
}
