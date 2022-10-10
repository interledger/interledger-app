package workflows

import (
	"time"

	"gitlab.com/fynbos/backend/providers/machnet"

	"go.temporal.io/sdk/workflow"
)

func CreateSendUserWorkflow(ctx workflow.Context, walletID string) (string, error) {
	var a *Activity
	ao := workflow.ActivityOptions{
		StartToCloseTimeout: 10 * time.Second,
	}
	ctx = workflow.WithActivityOptions(ctx, ao)

	logger := workflow.GetLogger(ctx)
	logger.Info("CreateSendUserWorkflow workflow started", "walletID", walletID)

	var externalUserID string
	err := workflow.ExecuteActivity(ctx, a.CreateExternalSendUser, walletID).Get(ctx, &externalUserID)
	if err != nil {
		logger.Error("CreateExternalSendUser Activity failed.", "Error", err)
		return "", err
	}

	err = workflow.ExecuteActivity(ctx, a.CreateUser, walletID, externalUserID).Get(ctx, nil)
	if err != nil {
		logger.Error("CreateExternalSendUser Activity failed.", "Error", err)
		return "", err
	}

	err = workflow.ExecuteActivity(ctx, a.StartExternalKYC, externalUserID).Get(ctx, nil)
	if err != nil {
		logger.Error("StartExternalKYC Activity failed.", "Error", err)
		return "", err
	}

	logger.Info("CreateSendUserWorkflow completed.", "external_user_id", externalUserID)

	return externalUserID, nil
}

func CreateTransactionWorkflow(ctx workflow.Context, args machnet.CreateTransactionArgs) (string, error) {
	var a *Activity
	ao := workflow.ActivityOptions{
		StartToCloseTimeout: 20 * time.Second,
	}
	ctx = workflow.WithActivityOptions(ctx, ao)

	logger := workflow.GetLogger(ctx)
	logger.Info("CreateTransactionWorkflow workflow started", "walletID", args.FromWalletID, "Amount", args.Amount)

	var trxID string
	err := workflow.ExecuteActivity(ctx, a.CreateTransaction, args).Get(ctx, &trxID)
	if err != nil {
		logger.Error("CreateTransaction Activity failed.", "Error", err)
		return "", err
	}

	err = workflow.ExecuteActivity(ctx, a.DeliverTransaction, args.FromWalletID, trxID).Get(ctx, nil)
	if err != nil {
		logger.Error("DeliverTransaction Activity failed.", "Error", err)
		return "", err
	}

	logger.Info("CreateTransactionWorkflow completed.", "external_transaction_id", trxID)

	return trxID, nil
}
