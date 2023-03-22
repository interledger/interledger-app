package ops

import (
	"gitlab.com/fynbos/backend/providers/gmt/external"
	"gitlab.com/fynbos/log"
	"time"

	"gitlab.com/fynbos/backend/providers/gmt"
	"go.temporal.io/sdk/workflow"
)

func OnboardUserWorkflow(ctx workflow.Context, walletID string) (string, error) {
	var a *Activity

	ao := workflow.ActivityOptions{
		StartToCloseTimeout: 20 * time.Minute,
	}
	ctx = workflow.WithActivityOptions(ctx, ao)

	logger := workflow.GetLogger(ctx)
	logger.Info("OnboardUserWorkflow workflow started", "walletID", walletID)

	err := workflow.ExecuteActivity(ctx, a.CheckWalletOFAC, walletID).Get(ctx, nil)
	if err != nil {
		logger.Error("failed to do OFAC checks", "err", err)
		return "", err
	}

	var cr ComplianceResp
	err = workflow.ExecuteActivity(ctx, a.IndividualCompliance, walletID).Get(ctx, &cr)
	if err != nil {
		logger.Error("failed to do compliance checks", "err", err)
		return "", err
	}

	err = workflow.ExecuteActivity(ctx, a.UpdateSendRecvUser, cr).Get(ctx, nil)
	if err != nil {
		logger.Error("failed to upsert gmt send recv user", "err", err)
		return "", err
	}

	return "TODO", nil
}

func ACH2ACHTransferWorkflow(ctx workflow.Context, args gmt.TransfersArgs) (string, error) {
	var a *Activity

	ao := workflow.ActivityOptions{
		StartToCloseTimeout: 20 * time.Minute,
	}
	ctx = workflow.WithActivityOptions(ctx, ao)

	logger := workflow.GetLogger(ctx)
	logger.Info("ACH2ACHTransferWorkflow workflow started", "From", args.FromLinkedAccountID, "Amount", args.Amount)

	err := workflow.ExecuteActivity(ctx, a.CheckAccountOFAC, args.ToLinkedAccountID).Get(ctx, nil)
	if err != nil {
		logger.Error("failed to do to linked account OFAC checks", "err", err)
		return "", err
	}

	err = workflow.ExecuteActivity(ctx, a.CheckAccountOFAC, args.FromLinkedAccountID).Get(ctx, nil)
	if err != nil {
		logger.Error("failed to do from linked account OFAC checks", "err", err)
		return "", err
	}

	var cr ComplianceResp
	err = workflow.ExecuteActivity(ctx, a.ACHCompliance, args).Get(ctx, &cr)
	if err != nil {
		logger.Error("failed to do compliance checks", "err", err)
		return "", err
	}

	err = workflow.ExecuteActivity(ctx, a.UpdateSendRecvUser, cr).Get(ctx, nil)
	if err != nil {
		logger.Error("failed to upsert gmt send recv user", "err", err)
		return "", err
	}

	var tr TransactionResp
	err = workflow.ExecuteActivity(ctx, a.InsertACH, args).Get(ctx, &tr)
	if err != nil {
		logger.Error("failed to insert gmt transaction", "err", err)
		return "", err
	}

	// TODO: Insert/update transactions

	err = workflow.ExecuteActivity(ctx, a.SaveReceipt, tr).Get(ctx, nil)
	if err != nil {
		logger.Error("failed to save gmt transaction receipt", "err", err)
		return "", err
	}

	// TODO: risk Scores if we want

	err = workflow.ExecuteActivity(ctx, a.VerifyTransaction, tr.ID).Get(ctx, nil)
	if err != nil {
		logger.Error("failed to verify gmt transaction", "err", err)
		return "", err
	}

	var refID string
	err = workflow.ExecuteActivity(ctx, a.CreateWorkflowRef, CreateWorkflowRefArgs{
		ExternalID:    tr.ID,
		WorkflowID:    workflow.GetInfo(ctx).WorkflowExecution.ID,
		WorkflowRunID: workflow.GetInfo(ctx).WorkflowExecution.RunID,
		ActivityName:  "ACH_to_ACH",
	}).Get(ctx, &refID)

	gmtChan := workflow.GetSignalChannel(ctx, gmtEventsChannel)
	for {
		var notify external.WsNotifications
		gmtChan.Receive(ctx, &notify)
		if notify.Password != tr.ID {
			log.Error("received notification for different transaction")
			continue
		}

		if notify.Status == external.TransactionStatusPaid {
			break
		}

		logger.Info("transaction status notification received", "id", notify.Password, "status", notify.Status)
		// TODO: handle edge cases
	}

	err = workflow.ExecuteActivity(ctx, a.CompleteWorkflowRef, refID).Get(ctx, nil)
	if err != nil {
		logger.Error("failed to complete workflow ref", "err", err)
		return "", err
	}

	return "TODO", nil
}
