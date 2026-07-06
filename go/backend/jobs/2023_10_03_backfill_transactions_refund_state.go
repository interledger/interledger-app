package jobs

import (
	"context"
	"fmt"
	"github.com/interledger/interledger-app/go/backend/db"
	"github.com/interledger/interledger-app/go/backend/transactions"
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

	var page db.Pagination
	for {
		var txs []transactions.Transaction
		err := workflow.ExecuteActivity(ctx, a.ListAllTransactions, page).Get(ctx, &txs)
		if err != nil {
			logger.Error("ListAllTransactions Activity failed.", "Error", err)
			return err
		}

		for i, tx := range txs {
			if i == page.PageSize {
				page.PageToken = tx.ID
				break
			}
			if tx.RefundState == transactions.RefundStateNone &&
				tx.State == transactions.StateFailed &&
				(tx.Type == transactions.TransactionTypeOpenOutgoingPayment || tx.Type == transactions.TransactionTypeSent) {
				logger.Info("Updating refund state for transaction", "tx", tx.ID)
				err := workflow.ExecuteActivity(ctx, a.UpdateTransactionRefundState, tx).Get(ctx, nil)
				if err != nil {
					logger.Error("UpdateTransactionRefundState Activity failed.", "Error", err)
					return err
				}
			}
		}

		if len(txs) <= page.PageSize {
			break
		}
	}

	return nil
}

func (a *Activity) ListAllTransactions(ctx context.Context, page db.Pagination) ([]transactions.Transaction, error) {
	return a.b.Transactions().ListAll(ctx, page)
}

func (a *Activity) UpdateTransactionRefundState(ctx context.Context, tx transactions.Transaction) error {
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
