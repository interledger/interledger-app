package jobs

import (
	"strings"
	"time"

	gmt_ops "gitlab.com/fynbos/backend/providers/gmt/ops"
	"gitlab.com/fynbos/backend/transactions"
	"go.temporal.io/sdk/workflow"
)

func ClearOrphanedGMTTransactions(ctx workflow.Context, txIDs string) error {
	var gmtActivity *gmt_ops.Activity

	ao := workflow.ActivityOptions{
		StartToCloseTimeout: 10 * time.Minute,
	}
	ctx = workflow.WithActivityOptions(ctx, ao)

	logger := workflow.GetLogger(ctx)

	for _, txID := range strings.Split(txIDs, ",") {
		// Notify GMT of completed card transaction. Their state machine goes from "created" -> "transmitted" -> "paid" so we need to do 2 updates
		err := workflow.ExecuteActivity(ctx, gmtActivity.UpdateCardTransactionStatus, txID, transactions.StatePending).Get(ctx, nil)
		if err != nil {
			logger.Warn("failed to update card transaction on gmt", "err", err, "tx_id", txID)
			return err
		}

		err = workflow.ExecuteActivity(ctx, gmtActivity.UpdateCardTransactionStatus, txID, transactions.StateCompleted).Get(ctx, nil)
		if err != nil {
			logger.Warn("failed to update card transaction on gmt", "err", err, "tx_id", txID)
			return err
		}
	}

	return nil
}
