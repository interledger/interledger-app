package ops

import (
	"context"
	"fmt"
	"time"

	"github.com/interledger/interledger-app/go/backend/providers/gatehub/external"
	"github.com/interledger/interledger-app/go/backend/slack"
	"github.com/interledger/interledger-app/go/backend/transactions"
	"github.com/interledger/interledger-app/go/log"

	"go.temporal.io/api/enums/v1"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"
)

func StartClearingCardTransactionsPolling(b Backends) {
	// Every day at 4PM UTC
	schedule := "0 16 * * *"
	workflowID := "gatehub_card_transactions_clearing_poll"
	workflowOptions := client.StartWorkflowOptions{
		ID:                    workflowID,
		TaskQueue:             "backend",
		CronSchedule:          schedule,
		WorkflowIDReusePolicy: enums.WORKFLOW_ID_REUSE_POLICY_TERMINATE_IF_RUNNING,
	}

	we, err := b.Temporal().ExecuteWorkflow(context.Background(), workflowOptions, GatehubClearingCardTransactionsPollWorkflow)
	if err != nil {
		log.Fatal("Unable to execute workflow", zap.Error(err))
	}
	log.Info("Started workflow", zap.String("WorkflowID", we.GetID()), zap.String("RunID", we.GetRunID()))
}

func GatehubClearingCardTransactionsPollWorkflow(ctx workflow.Context) error {
	var a *Activity
	ctx = workflow.WithActivityOptions(ctx, workflow.ActivityOptions{
		StartToCloseTimeout: 20 * time.Minute,
	})

	logger := workflow.GetLogger(ctx)
	logger.Info("starting gatehub clearing card transactions polling")

	var txs []pendingTransaction
	if err := workflow.ExecuteActivity(ctx, a.GetPendingClearingCardTransactions).Get(ctx, &txs); err != nil {
		return err
	}

	if len(txs) == 0 {
		return nil
	}

	loopCtx := workflow.WithActivityOptions(ctx, workflow.ActivityOptions{
		StartToCloseTimeout: 30 * time.Second,
		RetryPolicy:         &temporal.RetryPolicy{MaximumAttempts: 3},
	})
	notifyCtx := workflow.WithActivityOptions(ctx, workflow.ActivityOptions{
		StartToCloseTimeout: 10 * time.Second,
		RetryPolicy:         &temporal.RetryPolicy{MaximumAttempts: 1},
	})

	for _, tx := range txs {
		var cardTx *external.CardTransaction
		if err := workflow.ExecuteActivity(loopCtx, a.GetCardTransaction, tx.UserID, tx.ID).Get(loopCtx, &cardTx); err != nil {
			_ = workflow.ExecuteActivity(notifyCtx, slack.SendToChannelActivity, slack.ChannelError, "wallet-info-bot", fmt.Sprintf("!!! Failed to fetch card transaction from GateHub\nGateHub TX ID: %s\nGateHub User ID: %s", tx.ID, tx.UserID)).Get(notifyCtx, nil)
			continue
		}

		if cardTx.TxStatus == nil {
			_ = workflow.ExecuteActivity(notifyCtx, slack.SendToChannelActivity, slack.ChannelError, "wallet-info-bot", fmt.Sprintf("!!! Card transaction has no status\nGateHub TX ID: %s\nGateHub User ID: %s", tx.ID, tx.UserID)).Get(notifyCtx, nil)
			continue
		}

		var walletID string
		if err := workflow.ExecuteActivity(loopCtx, a.GetWalletID, tx.UserID).Get(loopCtx, &walletID); err != nil {
			logger.Error("Failed fetching wallet ID", "txID", tx.ID, "error", err)
			continue
		}

		var internalTx *transactions.Transaction
		if err := workflow.ExecuteActivity(loopCtx, a.GetGateHubTransactionByForeignID, walletID, tx.ID).Get(loopCtx, &internalTx); err != nil {
			logger.Error("Failed fetching internal transaction", "txID", tx.ID, "error", err)
			continue
		}

		var activityErr error
		switch *cardTx.TxStatus {
		case external.CardTransactionStatusCompleted:
			activityErr = workflow.ExecuteActivity(loopCtx, a.FinalizeGatehubCardTransaction, tx.ID, internalTx.ID).Get(loopCtx, nil)
		case external.CardTransactionStatusFailed:
			activityErr = workflow.ExecuteActivity(loopCtx, a.RollbackGatehubCardTransaction, tx.ID, internalTx.ID).Get(loopCtx, nil)
		}
		if activityErr != nil {
			logger.Error("Failed updating card transaction", "txID", tx.ID, "error", activityErr)
		}

		workflow.Sleep(ctx, 100*time.Millisecond)
	}

	return nil
}

func StartRealtimeCardTransactionsPolling(b Backends) {
	// Every 2 minutes
	schedule := "*/2 * * * *"
	workflowID := "gatehub_card_transactions_realtime_poll"
	workflowOptions := client.StartWorkflowOptions{
		ID:                    workflowID,
		TaskQueue:             "backend",
		CronSchedule:          schedule,
		WorkflowIDReusePolicy: enums.WORKFLOW_ID_REUSE_POLICY_TERMINATE_IF_RUNNING,
	}

	we, err := b.Temporal().ExecuteWorkflow(context.Background(), workflowOptions, GatehubRealtimeCardTransactionsPollWorkflow)
	if err != nil {
		log.Fatal("Unable to execute workflow", zap.Error(err))
	}
	log.Info("Started workflow", zap.String("WorkflowID", we.GetID()), zap.String("RunID", we.GetRunID()))
}

func GatehubRealtimeCardTransactionsPollWorkflow(ctx workflow.Context) error {
	var a *Activity
	ctx = workflow.WithActivityOptions(ctx, workflow.ActivityOptions{
		StartToCloseTimeout: 4 * time.Minute,
	})

	logger := workflow.GetLogger(ctx)
	logger.Info("starting gatehub realtime card transactions polling")

	var txs []pendingTransaction
	if err := workflow.ExecuteActivity(ctx, a.GetPendingRealtimeCardTransactions).Get(ctx, &txs); err != nil {
		return err
	}

	if len(txs) == 0 {
		return nil
	}

	loopCtx := workflow.WithActivityOptions(ctx, workflow.ActivityOptions{
		StartToCloseTimeout: 30 * time.Second,
		RetryPolicy:         &temporal.RetryPolicy{MaximumAttempts: 3},
	})
	notifyCtx := workflow.WithActivityOptions(ctx, workflow.ActivityOptions{
		StartToCloseTimeout: 10 * time.Second,
		RetryPolicy:         &temporal.RetryPolicy{MaximumAttempts: 1},
	})

	for _, tx := range txs {
		var cardTx *external.CardTransaction
		if err := workflow.ExecuteActivity(loopCtx, a.GetCardTransaction, tx.UserID, tx.ID).Get(loopCtx, &cardTx); err != nil {
			_ = workflow.ExecuteActivity(notifyCtx, slack.SendToChannelActivity, slack.ChannelError, "wallet-info-bot", fmt.Sprintf("!!! Failed to fetch card transaction from GateHub\nGateHub TX ID: %s\nGateHub User ID: %s", tx.ID, tx.UserID)).Get(notifyCtx, nil)
			continue
		}

		if cardTx.TxStatus == nil {
			_ = workflow.ExecuteActivity(notifyCtx, slack.SendToChannelActivity, slack.ChannelError, "wallet-info-bot", fmt.Sprintf("!!! Card transaction has no status\nGateHub TX ID: %s\nGateHub User ID: %s", tx.ID, tx.UserID)).Get(notifyCtx, nil)
			continue
		}

		isMasterCardSend := cardTx.Type == external.CardTransactionTypeTransferFromAccount || cardTx.Type == external.CardTransactionTypeTransferToAccount
		isReversal := cardTx.TransactionClassification != nil && *cardTx.TransactionClassification == external.CardTransactionClassificationReversal
		if isMasterCardSend && isReversal {
			_ = workflow.ExecuteActivity(notifyCtx, slack.SendToChannelActivity, slack.ChannelError, "wallet-info-bot", fmt.Sprintf("!!! Unexpected reversal for MasterCard transfer card transaction\nGateHub TX ID: %s\nGateHub User ID: %s\nType: %d", tx.ID, tx.UserID, cardTx.Type)).Get(notifyCtx, nil)
			continue
		}

		var walletID string
		if err := workflow.ExecuteActivity(loopCtx, a.GetWalletID, tx.UserID).Get(loopCtx, &walletID); err != nil {
			logger.Error("Failed fetching wallet ID", "txID", tx.ID, "error", err)
			continue
		}

		var internalTx *transactions.Transaction
		if err := workflow.ExecuteActivity(loopCtx, a.GetGateHubTransactionByForeignID, walletID, tx.ID).Get(loopCtx, &internalTx); err != nil {
			logger.Error("Failed fetching internal transaction", "txID", tx.ID, "error", err)
			continue
		}

		var activityErr error
		if isReversal {
			if cardTx.RefTransactionID == nil {
				_ = workflow.ExecuteActivity(notifyCtx, slack.SendToChannelActivity, slack.ChannelError, "wallet-info-bot", fmt.Sprintf("!!! Reversal card transaction has no ref transaction id\nGateHub TX ID: %s\nGateHub User ID: %s", tx.ID, tx.UserID)).Get(notifyCtx, nil)
				continue
			}
			var originalInternalTx *transactions.Transaction
			if err := workflow.ExecuteActivity(loopCtx, a.GetGateHubTransactionByForeignID, walletID, *cardTx.RefTransactionID).Get(loopCtx, &originalInternalTx); err != nil {
				logger.Error("Failed fetching original internal transaction for reversal", "txID", tx.ID, "refTxID", *cardTx.RefTransactionID, "error", err)
				continue
			}
			switch *cardTx.TxStatus {
			case external.CardTransactionStatusCompleted:
				activityErr = workflow.ExecuteActivity(loopCtx, a.FinalizeGatehubCardReversal, FinalizeGatehubCardReversalArgs{
					ReversalCardTxID:     tx.ID,
					ReversalInternalTxID: internalTx.ID,
					OriginalInternalTxID: originalInternalTx.ID,
				}).Get(loopCtx, nil)
			case external.CardTransactionStatusFailed:
				activityErr = workflow.ExecuteActivity(loopCtx, a.FailGatehubCardReversal, tx.ID, internalTx.ID).Get(loopCtx, nil)
			}
		} else {
			switch *cardTx.TxStatus {
			case external.CardTransactionStatusCompleted:
				activityErr = workflow.ExecuteActivity(loopCtx, a.FinalizeGatehubCardTransaction, tx.ID, internalTx.ID).Get(loopCtx, nil)
			case external.CardTransactionStatusFailed:
				activityErr = workflow.ExecuteActivity(loopCtx, a.RollbackGatehubCardTransaction, tx.ID, internalTx.ID).Get(loopCtx, nil)
			}
		}
		if activityErr != nil {
			logger.Error("Failed updating card transaction", "txID", tx.ID, "error", activityErr)
		}
	}

	return nil
}
