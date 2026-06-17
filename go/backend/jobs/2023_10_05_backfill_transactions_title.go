package jobs

import (
	"context"
	"github.com/interledger/interledger-app/go/backend/db"
	"github.com/interledger/interledger-app/go/backend/transactions"
	ops_transactions "github.com/interledger/interledger-app/go/backend/transactions/ops"
	"go.temporal.io/sdk/workflow"
	"time"
)

func BackfillTransactionsTitle(ctx workflow.Context) error {
	var a *Activity
	ao := workflow.ActivityOptions{
		StartToCloseTimeout: 10 * time.Minute,
	}
	ctx = workflow.WithActivityOptions(ctx, ao)
	logger := workflow.GetLogger(ctx)

	var page db.Pagination
	page.PageSize = 50
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

			logger.Info("Updating title for transaction", "tx", tx.ID)
			err := workflow.ExecuteActivity(ctx, a.UpdateTransactionTitle, tx).Get(ctx, nil)
			if err != nil {
				logger.Error("UpdateTransactionTitle Activity failed.", "Error", err)
				return err
			}
		}

		if len(txs) <= page.PageSize {
			break
		}
	}

	return nil
}

// this runs a bit slow, we can make it faster by batching the updates if we use this in the future
func (a *Activity) UpdateTransactionTitle(ctx context.Context, tx transactions.Transaction) error {
	title := ops_transactions.GenerateTransactionTitle(ctx, a.b.Wallets(), ops_transactions.GenerateTransactionTitleArgs{
		Source:              tx.Source,
		Destination:         tx.Destination,
		Type:                tx.Type,
		DestinationIdentity: tx.DestinationIdentity,
	})

	_, err := a.b.DB().ExecContext(ctx, `UPDATE transactions SET title = $1 WHERE id = $2`, title, tx.ID)
	return err
}
