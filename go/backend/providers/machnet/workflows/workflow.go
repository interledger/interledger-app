package workflows

import (
	"fmt"
	"time"

	"gitlab.com/fynbos/backend/providers/machnet/ops"

	"gitlab.com/fynbos/backend/providers/machnet"
	"gitlab.com/fynbos/backend/providers/machnet/external"
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
	err := workflow.ExecuteActivity(ctx, a.UpsertExternalSendUser, walletID).Get(ctx, &externalUserID)
	if err != nil {
		logger.Error("UpsertExternalSendUser Activity failed.", "Error", err)
		return "", err
	}

	err = workflow.ExecuteActivity(ctx, a.CreateUser, walletID, externalUserID).Get(ctx, nil)
	if err != nil {
		logger.Error("CreateUser Activity failed.", "Error", err)
		return "", err
	}

	err = workflow.ExecuteActivity(ctx, a.StartExternalKYC, externalUserID).Get(ctx, nil)
	if err != nil {
		logger.Error("StartExternalKYC Activity failed.", "Error", err, "externalID", externalUserID)
		return "", err
	}

	// Wait for KYC passing
	workflowArgs := CreateUserWorkflowRefArgs{
		ExternalUserID: externalUserID,
		WorkflowID:     workflow.GetInfo(ctx).WorkflowExecution.ID,
		WorkflowRunID:  workflow.GetInfo(ctx).WorkflowExecution.RunID,
		ActivityName:   "StartExternalKYC",
	}
	err = workflow.ExecuteActivity(ctx, a.CreateUserWorkflowRef, workflowArgs).Get(ctx, nil)
	if err != nil {
		logger.Error("CreateUserWorkflowRef Activity failed.", "Error", err)
		return "", err
	}

	trxChan := workflow.GetSignalChannel(ctx, ops.UserEventsChannel)
	for {
		var kycEvent external.Event
		trxChan.Receive(ctx, &kycEvent)
		logger.Info("status event: external user ID=", kycEvent.UserID, "status=", kycEvent.EventName)
		if kycEvent.UserID != externalUserID { // not for this user
			logger.Error("Received notification for different user.")
			continue
		}

		if external.UserKYCVerified == kycEvent.EventName {
			break
		}

		if kycEvent.EventName == external.UserKYCRetry ||
			kycEvent.EventName == external.UserKYCSuspended {
			return "", temporal.NewNonRetryableApplicationError(fmt.Sprintf("user (%s) KYC failed (%s)", externalUserID, kycEvent.EventName), "ErrInternal", external.ErrInternal)
		}
	}

	err = workflow.ExecuteActivity(ctx, a.CompleteUserWorkflowRef, workflowArgs).Get(ctx, nil)
	if err != nil {
		logger.Error("CompleteUserWorkflowRef Activity failed.", "Error", err)
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
	}
	ctx = workflow.WithActivityOptions(ctx, ao)

	logger := workflow.GetLogger(ctx)
	logger.Info("CreateTransactionWorkflow workflow started", "From", args.FromLinkedAccountID, "To", args.ToLinkedAccountID, "Amount", args.Amount)

	var fundWallet bool
	err := workflow.ExecuteActivity(ctx, a.ShouldFundWallet, args).Get(ctx, &fundWallet)
	if err != nil {
		logger.Error("FundUserWalletFromCard Activity failed.", "Error", err)
		return "", err
	}

	// The fund wallet has 7 days to complete
	timeoutFuture := workflow.NewTimer(ctx, time.Hour*24*8)

	trxChan := workflow.GetSignalChannel(ctx, ops.TransactionEventsChannel)

	var doBreak bool
	var errToReturn error

	var fundWalletTX FundWalletResponse
	if fundWallet {
		err = workflow.ExecuteActivity(ctx, a.FundUserWalletFromCard, FundWalletArgs{
			CreateTransactionArgs: args,
			WorkflowID:            workflow.GetInfo(ctx).WorkflowExecution.ID,
		}).Get(ctx, &fundWalletTX)
		if err != nil {
			logger.Error("FundUserWalletFromCard Activity failed.", "Error", err)
			return "", err
		}

		err = workflow.ExecuteActivity(ctx, a.CreateTransactionWorkflowRef, CreateTransactionWorkflowRefArgs{
			FromLinkedAccountID:   args.FromLinkedAccountID,
			ExternalTransactionID: fundWalletTX.FundTX,
			WorkflowID:            workflow.GetInfo(ctx).WorkflowExecution.ID,
			WorkflowRunID:         workflow.GetInfo(ctx).WorkflowExecution.RunID,
			AcitivityName:         "FundUserWalletFromCard",
		}).Get(ctx, nil)
		if err != nil {
			logger.Error("CreateTransactionWorkflowRef Activity failed.", "Error", err)
			return "", err
		}

		for {
			if doBreak {
				break
			}
			selector := workflow.NewSelector(ctx)
			selector.AddFuture(timeoutFuture, func(f workflow.Future) {
				doBreak = true
				errToReturn = temporal.NewNonRetryableApplicationError("fund user wallet transaction has timed out", "ErrTimeout", machnet.ErrInternal)
			})

			selector.AddReceive(trxChan, func(c workflow.ReceiveChannel, _ bool) {
				var transactionEvent external.Event
				trxChan.Receive(ctx, &transactionEvent)
				logger.Info("status event: transactionID=", transactionEvent.ResourceID, "status=", transactionEvent.EventName)
				if transactionEvent.ResourceID != fundWalletTX.FundTX { // not for this transaction
					logger.Error("Received notification for different transaction.")
					return
				}

				if external.TransactionProcessedEvent == transactionEvent.EventName {
					doBreak = true
					return
				}

				if transactionEvent.EventName == external.TransactionFailedEvent ||
					transactionEvent.EventName == external.TransactionCancelledEvent {
					doBreak = true
					errToReturn = temporal.NewNonRetryableApplicationError(fmt.Sprintf("fund user wallet transaction failed event(%s)", transactionEvent.EventName), "ErrInternal", external.ErrInternal)
				}
			})

			// Wait the timer or the transaction to complete on machnet side.
			selector.Select(ctx)
		}

		if errToReturn != nil {
			return "", errToReturn
		}
	}

	var transferID string
	err = workflow.ExecuteActivity(
		ctx,
		a.StartWalletTransfer,
		StartWalletTransferArgs{
			CreateTransactionArgs: args,
			WorkflowID:            workflow.GetInfo(ctx).WorkflowExecution.ID,
			FundingTx:             fundWalletTX,
		},
	).Get(ctx, &transferID)
	if err != nil {
		logger.Error("StartWalletTransfer Activity failed.", "Error", err)
		return "", err
	}

	err = workflow.ExecuteActivity(ctx, a.CreateTransactionWorkflowRef, CreateTransactionWorkflowRefArgs{
		FromLinkedAccountID:   fundWalletTX.FromWalletLinkedAcc,
		ExternalTransactionID: transferID,
		WorkflowID:            workflow.GetInfo(ctx).WorkflowExecution.ID,
		WorkflowRunID:         workflow.GetInfo(ctx).WorkflowExecution.RunID,
		AcitivityName:         "StartWalletTransfer",
	}).Get(ctx, nil)
	if err != nil {
		logger.Error("CreateTransactionWorkflowRef Activity failed.", "Error", err)
		return "", err
	}

	// Wait for webhook to say if transfer is successful
	doBreak = false
	for {
		if doBreak {
			break
		}
		selector := workflow.NewSelector(ctx)
		selector.AddFuture(timeoutFuture, func(f workflow.Future) {
			doBreak = true
			errToReturn = temporal.NewNonRetryableApplicationError("wallet to wallet transaction has timed out", "ErrTimeout", machnet.ErrInternal)
		})

		selector.AddReceive(trxChan, func(c workflow.ReceiveChannel, _ bool) {
			var transactionEvent external.Event
			trxChan.Receive(ctx, &transactionEvent)
			logger.Info("status event: transactionID=", transactionEvent.ResourceID, "status=", transactionEvent.EventName)
			if transactionEvent.ResourceID != transferID { // not for this transaction
				logger.Error("Received notification for different transaction.")
				return
			}

			if external.TransactionProcessedEvent == transactionEvent.EventName {
				doBreak = true
				return
			}

			if transactionEvent.EventName == external.TransactionFailedEvent ||
				transactionEvent.EventName == external.TransactionCancelledEvent {
				doBreak = true
				errToReturn = temporal.NewNonRetryableApplicationError(fmt.Sprintf("wallet transfer failed failed event(%s)", transactionEvent.EventName), "ErrInternal", external.ErrInternal)
			}
		})

		// Wait the timer or the transaction to complete on machnet side.
		selector.Select(ctx)
	}

	if errToReturn != nil {
		return "", errToReturn
	}

	logger.Info("CreateTransactionWorkflow completed.", "fund_transfer_id", fundWalletTX.FundTX, "external_transfer_id", transferID)

	return transferID, nil
}

func DeleteAccountWorkflow(ctx workflow.Context, linkedAccountID string) error {
	var a *Activity
	ao := workflow.ActivityOptions{
		StartToCloseTimeout: 20 * time.Minute,
	}

	ctx = workflow.WithActivityOptions(ctx, ao)

	logger := workflow.GetLogger(ctx)
	logger.Info("DeleteAccountWorkflow workflow started", "linkedAccountID", linkedAccountID)

	err := workflow.ExecuteActivity(ctx, a.DeleteUserFundSource, linkedAccountID).Get(ctx, nil)
	if err != nil {
		logger.Error("DeleteUserFundSource Activity failed.", "Error", err)
		return err
	}

	err = workflow.ExecuteActivity(ctx, a.DeleteLinkedAccount, linkedAccountID).Get(ctx, nil)
	if err != nil {
		logger.Error("DeleteLinkedAccount Activity failed.", "Error", err)
		return err
	}

	return nil
}
