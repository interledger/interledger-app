package ops

import (
	"context"
	"fmt"
	"time"

	"gitlab.com/fynbos/backend/providers/gatehub/external"
	"gitlab.com/fynbos/backend/slack"
	"gitlab.com/fynbos/log"

	"go.temporal.io/api/enums/v1"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"
)

func StartCardTransactionsPooling(b Backends) {
	// Every day at 4PM UTC
	schedule := "0 16 * * *"
	workflowID := "gatehub_card_transactions_poll"
	workflowOptions := client.StartWorkflowOptions{
		ID:                    workflowID,
		TaskQueue:             "backend",
		CronSchedule:          schedule,
		WorkflowIDReusePolicy: enums.WORKFLOW_ID_REUSE_POLICY_TERMINATE_IF_RUNNING,
	}

	we, err := b.Temporal().ExecuteWorkflow(context.Background(), workflowOptions, GatehubCardTransactionsPollWorkflow)
	if err != nil {
		log.Fatal("Unable to execute workflow", zap.Error(err))
	}
	log.Info("Started workflow", zap.String("WorkflowID", we.GetID()), zap.String("RunID", we.GetRunID()))
}

func GatehubCardTransactionsPollWorkflow(ctx workflow.Context) error {
	var a *Activity
	ao := workflow.ActivityOptions{
		StartToCloseTimeout: 20 * time.Minute,
	}
	ctx = workflow.WithActivityOptions(ctx, ao)

	logger := workflow.GetLogger(ctx)

	logger.Info("starting gatehub card transactions polling")

	var txs []pendingTransaction
	err := workflow.ExecuteActivity(ctx, a.GetPendingCardTransactions).Get(ctx, &txs)
	if err != nil {
		return err
	}

	// No new transactions, so nothing to do
	if len(txs) == 0 {
		return nil
	}

	for _, tx := range txs {
		err := workflow.ExecuteActivity(ctx, a.ProcessCardTransaction, tx).Get(ctx, nil)
		if err != nil {
			logger.Error("Failed processing transaction", "txID", tx, "error", err)
			slack.SendToChannel(context.Background(), "wallet-info-bot", fmt.Sprintf(":warning: Failed to process card transaction\nGateHub TX ID: %s\nGateHub User ID: %s", tx.ID, tx.UserID))
			continue
		}

		time.Sleep(time.Millisecond * 300)
	}

	return nil
}

func (a *Activity) GetPendingCardTransactions(ctx context.Context) ([]pendingTransaction, error) {
	return GetPendingCardTransactions(ctx, a.b)
}

func (a *Activity) ProcessCardTransaction(ctx context.Context, tx pendingTransaction) error {
	walletID, err := getWalletID(ctx, a.b, tx.UserID)
	if err != nil {
		return err
	}

	cardTx, err := a.external.GetCardTransaction(ctx, tx.UserID, tx.ID)
	if err != nil {
		return err
	}

	internalTx, err := a.b.Transactions().GetTransactionByForeignID(ctx, walletID, tx.ID)
	if err != nil {
		return err
	}

	switch *cardTx.TxStatus {
	case external.CardTractionStatusCompleted:
		err = a.FinalizeGatehubCardTransaction(ctx, cardTx.TransactionID, internalTx.ID)
		if err != nil {
			return err
		}
	case external.CardTractionStatusFailed:
		err = a.RollbackGatehubCardTransaction(ctx, cardTx.TransactionID, internalTx.ID)
		if err != nil {
			return err
		}
	default:
		//do nothing
	}

	return nil
}
