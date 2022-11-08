package workflows

import (
	"fmt"
	"time"

	"gitlab.com/fynbos/backend/providers/machnet"
	"gitlab.com/fynbos/backend/providers/machnet/external"
	"gitlab.com/fynbos/backend/providers/machnet/ops"
	"go.temporal.io/sdk/temporal"
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

	err = workflow.ExecuteActivity(ctx, a.CreateWallet, externalUserID).Get(ctx, nil)
	if err != nil {
		logger.Error("CreateWallet Activity failed.", "Error", err)
		return "", err
	}

	logger.Info("CreateSendUserWorkflow completed.", "external_user_id", externalUserID)

	return externalUserID, nil
}

func CreateTransactionWorkflow(ctx workflow.Context, args machnet.CreateTransactionArgs) (string, error) {
	var a *Activity
	ao := workflow.ActivityOptions{
		StartToCloseTimeout: 20 * time.Minute,
		RetryPolicy: &temporal.RetryPolicy{
			MaximumAttempts: 30,
		},
	}
	ctx = workflow.WithActivityOptions(ctx, ao)

	logger := workflow.GetLogger(ctx)
	logger.Info("CreateTransactionWorkflow workflow started", "From", args.FromLinkedAccountID, "To", args.ToLinkedAccountID, "Amount", args.Amount)

	var fundWalletTX FundWalletResponse
	err := workflow.ExecuteActivity(ctx, a.FundUserWalletFromCard, args).Get(ctx, &fundWalletTX)
	if err != nil {
		logger.Error("FundUserWalletFromCard Activity failed.", "Error", err)
		return "", err
	}

	err = workflow.ExecuteActivity(ctx, a.CreateTransactionWorkflowRef, CreateTransactionWorkflowRefArgs{
		FromLinkedAccountID:   args.FromLinkedAccountID,
		ExternalTransactionID: fundWalletTX.FundTX,
		WorkflowID:            workflow.GetInfo(ctx).WorkflowExecution.ID,
		WorkflowRunID:         workflow.GetInfo(ctx).WorkflowExecution.RunID,
	}).Get(ctx, nil)
	if err != nil {
		logger.Error("CreateTransactionWorkflowRef Activity failed.", "Error", err)
		return "", err
	}

	trxChan := workflow.GetSignalChannel(ctx, ops.TransactionEventsChannel)
	for {
		var transactionEvent external.Event
		trxChan.Receive(ctx, &transactionEvent)
		logger.Info("status event: transactionID=", transactionEvent.ResourceID, "status=", transactionEvent.EventName)
		if transactionEvent.ResourceID != fundWalletTX.FundTX { // not for this transaction
			logger.Error("Received notification for different transaction.")
			continue
		}

		if external.TransactionProcessedEvent == transactionEvent.EventName {
			break
		}

		if transactionEvent.EventName == external.TransactionFailedEvent ||
			transactionEvent.EventName == external.TransactionCancelledEvent {
			return "", temporal.NewNonRetryableApplicationError(fmt.Sprintf("fund user wallet transaction failed event(%s)", transactionEvent.EventName), "ErrInternal", external.ErrInternal)
		}
	}

	var transferID string
	err = workflow.ExecuteActivity(ctx, a.StartWalletTransfer, args, fundWalletTX).Get(ctx, &transferID)
	if err != nil {
		logger.Error("StartWalletTransfer Activity failed.", "Error", err)
		return "", err
	}

	err = workflow.ExecuteActivity(ctx, a.CreateTransactionWorkflowRef, CreateTransactionWorkflowRefArgs{
		FromLinkedAccountID:   fundWalletTX.FromWalletLinkedAcc,
		ExternalTransactionID: transferID,
		WorkflowID:            workflow.GetInfo(ctx).WorkflowExecution.ID,
		WorkflowRunID:         workflow.GetInfo(ctx).WorkflowExecution.RunID,
	}).Get(ctx, nil)
	if err != nil {
		logger.Error("CreateTransactionWorkflowRef Activity failed.", "Error", err)
		return "", err
	}

	// Wait for webhook to say if transfer is successful
	for {
		var transactionEvent external.Event
		trxChan.Receive(ctx, &transactionEvent)
		logger.Info("status event: transactionID=", transactionEvent.ResourceID, "status=", transactionEvent.EventName)
		if transactionEvent.ResourceID != transferID { // not for this transaction
			logger.Error("Received notification for different transaction.")
			continue
		}

		if external.TransactionProcessedEvent == transactionEvent.EventName {
			break
		}

		if transactionEvent.EventName == external.TransactionFailedEvent ||
			transactionEvent.EventName == external.TransactionCancelledEvent {
			return "", temporal.NewNonRetryableApplicationError(fmt.Sprintf("wallet transfer failed failed event(%s)", transactionEvent.EventName), "ErrInternal", external.ErrInternal)
		}
	}

	logger.Info("CreateTransactionWorkflow completed.", "fund_transfer_id", fundWalletTX.FundTX, "external_transfer_id", transferID)

	return transferID, nil
}
