package jobs

import (
	"context"
	"fmt"
	"gitlab.com/fynbos/backend/transactions"
	"go.temporal.io/sdk/workflow"
	"time"
)

func BackfillTransactionsRefundState(ctx workflow.Context) error {
	var a *Activity
	ao := workflow.ActivityOptions{
		StartToCloseTimeout: 10 * time.Minute,
	}
	ctx = workflow.WithActivityOptions(ctx, ao)
	logger := workflow.GetLogger(ctx)

	var txs []Transaction
	err := workflow.ExecuteActivity(ctx, a.ListAllTransactions).Get(ctx, txs)
	if err != nil {
		logger.Error("ListAllTransactions Activity failed.", "Error", err)
		return err
	}

	for _, tx := range txs {
		if tx.RefundState == transactions.RefundStateNone &&
			tx.State == transactions.StateFailed &&
			(tx.Type == transactions.TransactionTypeOpenOutgoingPayment || tx.Type == transactions.TransactionTypeOutgoing) {
			logger.Info("Updating refund state for transaction", "tx", tx.ID)
			err := workflow.ExecuteActivity(ctx, a.UpdateTransactionRefundState, tx).Get(ctx, nil)
			if err != nil {
				logger.Error("UpdateTransactionRefundState Activity failed.", "Error", err)
				return err
			}
		}
	}

	return nil
}

type Transaction struct {
	ID          string                       `db:"id"`
	State       transactions.State           `db:"state"`
	Type        transactions.TransactionType `db:"type"`
	RefundState transactions.RefundState     `db:"refund_state"`
}

func (a *Activity) ListAllTransactions(ctx context.Context) ([]Transaction, error) {
	var txs []Transaction

	err := a.b.DB().SelectContext(ctx, &txs, `SELECT id, state, type, refund_state FROM transactions`)
	if err != nil {
		return nil, err
	}

	return txs, nil
}

func (a *Activity) UpdateTransactionRefundState(ctx context.Context, tx Transaction) error {
	xfers, err := a.b.Transactions().ListTransfers(ctx, tx.ID)
	if err != nil {
		return fmt.Errorf("%w %s", transactions.ErrInternal, err)
	}

	var refundState transactions.RefundState
	for _, xfer := range xfers {
		if xfer.Type == transactions.TransferTypeDebitCard && tx.RefundState == transactions.RefundStateNone {
			refundState = transactions.RefundStatePending
		}

		if xfer.Type == transactions.TransferTypeCreditCard && refundState == transactions.RefundStatePending {
			refundState = transactions.RefundStateCompleted
		}
	}

	if refundState != transactions.RefundStateNone {
		err := a.b.Transactions().SetTransactionRefundState(ctx, tx.ID, refundState)
		if err != nil {
			return fmt.Errorf("%w %s", transactions.ErrInternal, err)
		}
	}

	return nil
}
