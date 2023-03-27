package ops

import (
	"time"

	"github.com/google/uuid"
	"gitlab.com/fynbos/backend/providers/gmt"
	"gitlab.com/fynbos/backend/providers/gmt/external"
	temporal_utils "gitlab.com/fynbos/backend/temporal/utils"
	"gitlab.com/fynbos/backend/transactions"
	"gitlab.com/fynbos/log"
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

func ACH2ACHTransferWorkflow(ctx workflow.Context, args gmt.TransfersArgs) (*gmt.TransferResponse, error) {
	var a *Activity

	ao := workflow.ActivityOptions{
		StartToCloseTimeout: 20 * time.Minute,
	}
	ctx = workflow.WithActivityOptions(ctx, ao)

	logger := workflow.GetLogger(ctx)
	logger.Info("ACH2ACHTransferWorkflow workflow started", "From", args.FromLinkedAccountID, "Amount", args.Amount)

	err := workflow.ExecuteActivity(ctx, a.CheckWalletOFAC, args.ToWalletID).Get(ctx, nil)
	if err != nil {
		logger.Error("failed to do to linked account OFAC checks", "err", err)
		if temporal_utils.IsNonRetryableError(err) {
			return &gmt.TransferResponse{
				State: transactions.StateFailed,
			}, nil
		}
		return nil, err
	}

	err = workflow.ExecuteActivity(ctx, a.CheckWalletOFAC, args.FromWalletID).Get(ctx, nil)
	if err != nil {
		logger.Error("failed to do from linked account OFAC checks", "err", err)
		if temporal_utils.IsNonRetryableError(err) {
			return &gmt.TransferResponse{
				State: transactions.StateFailed,
			}, nil
		}
		return nil, err
	}

	var cr ComplianceResp
	err = workflow.ExecuteActivity(ctx, a.ACHCompliance, args).Get(ctx, &cr)
	if err != nil {
		logger.Error("failed to do compliance checks", "err", err)
		if temporal_utils.IsNonRetryableError(err) {
			return &gmt.TransferResponse{
				State: transactions.StateFailed,
			}, nil
		}
		return nil, err
	}

	err = workflow.ExecuteActivity(ctx, a.UpdateSendRecvUser, cr).Get(ctx, nil)
	if err != nil {
		logger.Error("failed to upsert gmt send recv user", "err", err)
		return nil, err
	}

	var tr TransactionResp
	err = workflow.ExecuteActivity(ctx, a.InsertACH, args).Get(ctx, &tr)
	if err != nil {
		logger.Error("failed to insert gmt transaction", "err", err)
		if temporal_utils.IsNonRetryableError(err) {
			return &gmt.TransferResponse{
				State: transactions.StateFailed,
			}, nil
		}
		return nil, err
	}

	// Insert/update transfers
	err = workflow.ExecuteActivity(ctx, a.AddTransactionTransfer, args.FromTransactionID, []transactions.TransferArgs{
		{
			LinkedAccountID: args.FromLinkedAccountID,
			Type:            transactions.TransferTypeCreditBankAccount,
			Amount:          args.Amount,
			State:           transactions.StatePending,
			ForeignID:       tr.ID,
		},
	}).Get(ctx, nil)
	if err != nil {
		logger.Error("error updating transaction transfer", "Error", err)
		return nil, err
	}

	// Insert incoming transfer
	var recvTrxID string
	err = workflow.SideEffect(ctx, func(ctx workflow.Context) interface{} {
		return uuid.NewString()
	}).Get(&recvTrxID)
	if err != nil {
		logger.Error("error generating transactionID as side effect", "Error", err)
		return nil, err
	}

	err = workflow.ExecuteActivity(ctx, a.AddTransaction, transactions.CreateTransactionArgs{
		ID:          recvTrxID,
		WalletID:    args.ToWalletID,
		ForeignID:   args.ToForeignID,
		ForeignType: transactions.TransactionTypeOpenPaymentIncoming,
		Provider:    transactions.ProviderOpenPayments,
		State:       transactions.StatePending,
		Source:      args.FromPaymentPointer,
		Destination: args.ToPaymentPointer,
		Amount:      args.Amount,
		Transfers: []transactions.TransferArgs{
			{
				ForeignID:       tr.ID,
				LinkedAccountID: args.ToLinkedAccountID,
				Type:            transactions.TransferTypeCreditBankAccount,
				Amount:          args.Amount,
				State:           transactions.StatePending,
			},
		},
	}).Get(ctx, nil)
	if err != nil {
		logger.Error("failed to add transaction for recipient", "err", err)
		if temporal_utils.IsNonRetryableError(err) {
			return &gmt.TransferResponse{
				State:      transactions.StateFailed,
				ExternalID: tr.ID,
			}, nil
		}
		return nil, err
	}

	err = workflow.ExecuteActivity(ctx, a.SaveReceipt, tr).Get(ctx, nil)
	if err != nil {
		logger.Error("failed to save gmt transaction receipt", "err", err)
		return nil, err
	}

	// TODO: risk Scores if we want

	err = workflow.ExecuteActivity(ctx, a.VerifyTransaction, tr.ID).Get(ctx, nil)
	if err != nil {
		logger.Error("failed to verify gmt transaction", "err", err)
		if temporal_utils.IsNonRetryableError(err) {
			return &gmt.TransferResponse{
				State:      transactions.StateFailed,
				ExternalID: tr.ID,
			}, nil
		}
		return nil, err
	}

	var refID string
	err = workflow.ExecuteActivity(ctx, a.CreateWorkflowRef, CreateWorkflowRefArgs{
		ExternalID:    tr.ID,
		WorkflowID:    workflow.GetInfo(ctx).WorkflowExecution.ID,
		WorkflowRunID: workflow.GetInfo(ctx).WorkflowExecution.RunID,
		ActivityName:  "ACH_to_ACH",
	}).Get(ctx, &refID)
	if err != nil {
		logger.Error("failed to create workflow reference", "err", err)
		return nil, err
	}

	gmtChan := workflow.GetSignalChannel(ctx, gmtEventsChannel)
	state := transactions.StateCompleted
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
		// TODO: handle edge cases and set state to some error state
	}

	err = workflow.ExecuteActivity(ctx, a.CompleteWorkflowRef, refID).Get(ctx, nil)
	if err != nil {
		logger.Error("failed to complete workflow ref", "err", err)
		return nil, err
	}

	// update send and receive transaction state.
	err = workflow.ExecuteActivity(ctx, a.UpdateTransferStateByType, args.FromTransactionID, args.FromWalletID, transactions.TransferTypeDebitBankAccount, state).Get(ctx, nil)
	if err != nil {
		logger.Error("failed to update transaction state", "error", err, "state", state)
		return nil, err
	}
	err = workflow.ExecuteActivity(ctx, a.UpdateTransferStateByType, recvTrxID, args.ToWalletID, transactions.TransferTypeCreditBankAccount, state).Get(ctx, nil)
	if err != nil {
		logger.Error("failed to update transaction state", "error", err, "state", state)
		return nil, err
	}

	err = workflow.ExecuteActivity(ctx, a.ConfirmNotification, tr.ID).Get(ctx, nil)
	if err != nil {
		logger.Error("failed to clear notification", "error", err, "ext ID", tr.ID)
		if temporal_utils.IsNonRetryableError(err) {
			return &gmt.TransferResponse{
				State:      transactions.StateFailed,
				ExternalID: tr.ID,
			}, nil
		}
		return nil, err
	}

	return &gmt.TransferResponse{
		State:      state,
		ExternalID: tr.ID,
	}, nil
}
