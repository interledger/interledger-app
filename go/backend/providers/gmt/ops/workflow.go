package ops

import (
	"fmt"
	"strings"
	"time"

	"gitlab.com/fynbos/backend/currency"

	"gitlab.com/fynbos/backend/providers"

	"github.com/google/uuid"
	"gitlab.com/fynbos/backend/providers/gmt/external"
	httplog "gitlab.com/fynbos/backend/providers/http"
	"gitlab.com/fynbos/backend/providers/tabapay"
	temporal_utils "gitlab.com/fynbos/backend/temporal/utils"
	"gitlab.com/fynbos/backend/transactions"
	"gitlab.com/fynbos/log"
	"go.temporal.io/sdk/temporal"
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
	err = workflow.ExecuteActivity(ctx, a.IndividualCompliance, providers.TransfersArgs{
		Amount:       currency.FromFloat64(1, currency.USD),
		FromWalletID: walletID}).Get(ctx, &cr)
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

func ACH2ACHTransferWorkflow(ctx workflow.Context, args providers.TransfersArgs) (*providers.TransferResponse, error) {
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
			return &providers.TransferResponse{
				Type:                  providers.GMTACH2ACH,
				OutgoingTransferState: transactions.StateFailed,
				IncomingTransferState: transactions.StateFailed,
			}, nil
		}
		return nil, err
	}

	err = workflow.ExecuteActivity(ctx, a.CheckWalletOFAC, args.FromWalletID).Get(ctx, nil)
	if err != nil {
		logger.Error("failed to do from linked account OFAC checks", "err", err)
		if temporal_utils.IsNonRetryableError(err) {
			return &providers.TransferResponse{
				Type:                  providers.GMTACH2ACH,
				OutgoingTransferState: transactions.StateFailed,
				IncomingTransferState: transactions.StateFailed,
			}, nil
		}
		return nil, err
	}

	var cr ComplianceResp
	err = workflow.ExecuteActivity(ctx, a.ACHCompliance, args).Get(ctx, &cr)
	if err != nil {
		logger.Error("failed to do compliance checks", "err", err)
		if temporal_utils.IsNonRetryableError(err) {
			return &providers.TransferResponse{
				Type:                  providers.GMTACH2ACH,
				OutgoingTransferState: transactions.StateFailed,
				IncomingTransferState: transactions.StateFailed,
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
			return &providers.TransferResponse{
				Type:                  providers.GMTACH2ACH,
				OutgoingTransferState: transactions.StateFailed,
				IncomingTransferState: transactions.StateFailed,
			}, nil
		}
		return nil, err
	}

	err = workflow.ExecuteActivity(ctx, a.SaveReceipt, tr).Get(ctx, nil)
	if err != nil {
		logger.Error("failed to save gmt transaction receipt", "err", err)
		return nil, err
	}

	txState := transactions.StatePending
	if strings.EqualFold(tr.Status, "Hold") {
		txState = transactions.StateOnHold
		err = workflow.ExecuteActivity(ctx, a.UpdateTransactionState, args.FromTransactionID, txState).Get(ctx, nil)
		if err != nil {
			logger.Error("error updating transaction state to on hold", "Error", err)
			return nil, err
		}
	}

	// Insert/update transfers
	err = workflow.ExecuteActivity(ctx, a.AddTransactionTransfer, args.FromTransactionID, []transactions.TransferArgs{
		{
			LinkedAccountID: args.FromLinkedAccountID,
			Type:            transactions.TransferTypeDebitBankAccount,
			Amount:          args.Amount,
			State:           txState,
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
		State:       txState,
		Source:      args.FromPaymentPointer,
		Destination: args.ToPaymentPointer,
		Amount:      args.Amount,
		Transfers: []transactions.TransferArgs{
			{
				ForeignID:       tr.ID,
				LinkedAccountID: args.ToLinkedAccountID,
				Type:            transactions.TransferTypeCreditBankAccount,
				Amount:          args.Amount,
				State:           txState,
			},
		},
	}).Get(ctx, nil)
	if err != nil {
		logger.Error("failed to add transaction for recipient", "err", err)
		if temporal_utils.IsNonRetryableError(err) {
			return &providers.TransferResponse{
				Type:                       providers.GMTACH2ACH,
				OutgoingTransferState:      transactions.StateFailed,
				OutgoingTransferExternalID: tr.ID,
				IncomingTransferState:      transactions.StateFailed,
				IncomingTransferExternalID: tr.ID,
			}, nil
		}
		return nil, err
	}

	// TODO: risk Scores if we want

	err = workflow.ExecuteActivity(ctx, a.VerifyTransaction, tr.ID).Get(ctx, nil)
	if err != nil {
		logger.Error("failed to verify gmt transaction", "err", err)
		if temporal_utils.IsNonRetryableError(err) {
			return &providers.TransferResponse{
				Type:                       providers.GMTACH2ACH,
				OutgoingTransferState:      transactions.StateFailed,
				OutgoingTransferExternalID: tr.ID,
				IncomingTransferState:      transactions.StateFailed,
				IncomingTransferExternalID: tr.ID,
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

		// If the transaction was OnHold then gets Released so it's back to pending
		if notify.Status == external.TransactionStatusReleased {
			// update send and receive transfer state.
			err = workflow.ExecuteActivity(ctx, a.UpdateTransactionState, args.FromTransactionID, transactions.StatePending).Get(ctx, nil)
			if err != nil {
				logger.Error("error updating transaction state to pending", "Error", err)
				return nil, err
			}

			err = workflow.ExecuteActivity(ctx, a.UpdateTransactionState, recvTrxID, transactions.StatePending).Get(ctx, nil)
			if err != nil {
				logger.Error("error updating transaction state to pending", "Error", err)
				return nil, err
			}

			err = workflow.ExecuteActivity(ctx, a.UpdateTransferStateByType, args.FromTransactionID, args.FromWalletID, transactions.TransferTypeDebitBankAccount, transactions.StatePending).Get(ctx, nil)
			if err != nil {
				logger.Error("failed to update transaction state", "error", err, "state", transactions.StatePending)
				return nil, err
			}
			err = workflow.ExecuteActivity(ctx, a.UpdateTransferStateByType, recvTrxID, args.ToWalletID, transactions.TransferTypeCreditBankAccount, transactions.StatePending).Get(ctx, nil)
			if err != nil {
				logger.Error("failed to update transaction state", "error", err, "state", transactions.StatePending)
				return nil, err
			}
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

	// update send and receive transfer state.
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

	err = workflow.ExecuteActivity(ctx, a.ConfirmPaidNotification, tr.ID).Get(ctx, nil)
	if err != nil {
		logger.Error("failed to clear paid notification", "error", err, "ext ID", tr.ID)
		if temporal_utils.IsNonRetryableError(err) {
			return &providers.TransferResponse{
				Type:                       providers.GMTACH2ACH,
				OutgoingTransferState:      transactions.StateFailed,
				OutgoingTransferExternalID: tr.ID,
				IncomingTransferState:      transactions.StateFailed,
				IncomingTransferExternalID: tr.ID,
			}, nil
		}
		return nil, err
	}

	return &providers.TransferResponse{
		Type:                       providers.GMTACH2ACH,
		OutgoingTransferState:      state,
		OutgoingTransferExternalID: tr.ID,
		IncomingTransferState:      state,
		IncomingTransferExternalID: tr.ID,
	}, nil
}

func Card2ACHTransferWorkflow(ctx workflow.Context, args providers.TransfersArgs) (*providers.TransferResponse, error) {
	var a *Activity

	ao := workflow.ActivityOptions{
		StartToCloseTimeout: 20 * time.Minute,
	}
	ctx = workflow.WithActivityOptions(ctx, ao)

	logger := workflow.GetLogger(ctx)
	logger.Info("Card2ACHTransferWorkflow workflow started", "From", args.FromLinkedAccountID, "Amount", args.Amount)

	err := workflow.ExecuteActivity(ctx, a.CheckWalletOFAC, args.ToWalletID).Get(ctx, nil)
	if err != nil {
		logger.Error("failed to do to linked account OFAC checks", "err", err)
		if temporal_utils.IsNonRetryableError(err) {
			return &providers.TransferResponse{
				Type:                  providers.GMTCARD2ACH,
				OutgoingTransferState: transactions.StateFailed,
				IncomingTransferState: transactions.StateFailed,
			}, nil
		}
		return nil, err
	}

	err = workflow.ExecuteActivity(ctx, a.CheckWalletOFAC, args.FromWalletID).Get(ctx, nil)
	if err != nil {
		logger.Error("failed to do from linked account OFAC checks", "err", err)
		if temporal_utils.IsNonRetryableError(err) {
			return &providers.TransferResponse{
				Type:                  providers.GMTCARD2ACH,
				OutgoingTransferState: transactions.StateFailed,
				IncomingTransferState: transactions.StateFailed,
			}, nil
		}
		return nil, err
	}

	var cr ComplianceResp
	err = workflow.ExecuteActivity(ctx, a.Card2ACHCompliance, args).Get(ctx, &cr)
	if err != nil {
		logger.Error("failed to do compliance checks", "err", err)
		if temporal_utils.IsNonRetryableError(err) {
			return &providers.TransferResponse{
				Type:                  providers.GMTCARD2ACH,
				OutgoingTransferState: transactions.StateFailed,
				IncomingTransferState: transactions.StateFailed,
			}, nil
		}
		return nil, err
	}

	err = workflow.ExecuteActivity(ctx, a.UpdateSendRecvUser, cr).Get(ctx, nil)
	if err != nil {
		logger.Error("failed to upsert gmt send recv user", "err", err)
		return &providers.TransferResponse{
			Type:                  providers.GMTCARD2ACH,
			OutgoingTransferState: transactions.StateFailed,
			IncomingTransferState: transactions.StateFailed,
		}, nil
	}

	var tabapayReferenceID string
	err = workflow.SideEffect(ctx, func(ctx workflow.Context) interface{} {
		return tabapay.NewReferenceID()
	}).Get(&tabapayReferenceID)
	if err != nil {
		logger.Error("error generating tabapay ReferenceID as side effect", "err", err)
		return &providers.TransferResponse{
			Type:                  providers.GMTCARD2ACH,
			OutgoingTransferState: transactions.StateFailed,
			IncomingTransferState: transactions.StateFailed,
		}, nil
	}

	newCtx := workflow.WithValue(ctx, httplog.ContextKey, &httplog.Metadata{
		Context: fmt.Sprintf("transactionID=%s", args.FromTransactionID),
	})
	newCtx = workflow.WithActivityOptions(newCtx, workflow.ActivityOptions{
		StartToCloseTimeout: 95 * time.Second, // cmay take up to 90 seconds.
		RetryPolicy: &temporal.RetryPolicy{
			MaximumAttempts: 2, // tabapay will block us if we keep retrying
		},
	})
	var tabapayTransaction tabapay.Transaction
	err = workflow.ExecuteActivity(newCtx, a.PullFromCard, PullFromCardArgs{
		TransactionID:       args.FromTransactionID,
		CardLinkedAccountID: args.FromLinkedAccountID,
		Amount:              args.Amount,
		ThreeDSID:           args.ThreeDSID,
		ReferenceID:         tabapayReferenceID,
	}).Get(ctx, &tabapayTransaction)
	if err != nil {
		logger.Error("failed to pull from card", "err", err)
		if temporal_utils.IsNonRetryableError(err) {
			return &providers.TransferResponse{
				Type:                       providers.GMTCARD2ACH,
				OutgoingTransferState:      transactions.StateFailed,
				OutgoingTransferExternalID: tabapayTransaction.ID,
				IncomingTransferState:      transactions.StateFailed,
			}, nil
		}
		return nil, err
	}
	if tabapay.IsTransactionStatusUnknown(tabapayTransaction) {
		// check again in 90 sec
		_ = workflow.Sleep(ctx, time.Second*90)
		logger.Info("Tabapay transaction status unknown. Checking again. id=", tabapayTransaction.ID)
		err = workflow.ExecuteActivity(newCtx, a.GetTabapayTransaction, tabapayTransaction.ID).Get(newCtx, &tabapayTransaction)
	}
	if err != nil || !tabapay.IsSuccessfulTransaction(tabapayTransaction) {
		logger.Error("failed to pull from card", "err", err)
		if temporal_utils.IsNonRetryableError(err) {
			return &providers.TransferResponse{
				Type:                       providers.GMTCARD2ACH,
				OutgoingTransferState:      transactions.StateFailed,
				OutgoingTransferExternalID: tabapayTransaction.ID,
				IncomingTransferState:      transactions.StateFailed,
			}, nil
		}
		return nil, err
	}

	// TODO: risk assessment

	// Insert/update transfers
	err = workflow.ExecuteActivity(ctx, a.AddTransactionTransfer, args.FromTransactionID, []transactions.TransferArgs{
		{
			LinkedAccountID: args.FromLinkedAccountID,
			Type:            transactions.TransferTypeDebitCard,
			Amount:          args.Amount,
			State:           transactions.StateCompleted,
			ForeignID:       tabapayTransaction.ID,
		},
	}).Get(ctx, nil)
	if err != nil {
		logger.Error("error updating transaction transfer", "Error", err)
		return nil, err
	}

	// Insert ACH
	var achTransaction TransactionResp
	err = workflow.ExecuteActivity(ctx, a.InsertCard2ACH, args).Get(ctx, &achTransaction)
	if err != nil {
		logger.Error("failed to insert gmt transaction", "err", err)
		if temporal_utils.IsNonRetryableError(err) {
			err = workflow.ExecuteActivity(newCtx, a.ReverseTabapayTransaction, tabapayTransaction.ID).Get(newCtx, nil)
			if err != nil {
				logger.Error("failed to rollback tabapay card transaction", "err", err)
			}
			return &providers.TransferResponse{
				Type:                       providers.GMTCARD2ACH,
				OutgoingTransferState:      transactions.StateCompleted,
				OutgoingTransferExternalID: tabapayTransaction.ID,
				IncomingTransferState:      transactions.StateFailed,
			}, nil
		}
		return nil, err
	}

	err = workflow.ExecuteActivity(ctx, a.SaveReceipt, achTransaction).Get(ctx, nil)
	if err != nil {
		logger.Error("failed to save gmt transaction receipt", "err", err)
		return nil, err
	}

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
				ForeignID:       achTransaction.ID,
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
			err = workflow.ExecuteActivity(newCtx, a.ReverseTabapayTransaction, tabapayTransaction.ID).Get(newCtx, nil)
			if err != nil {
				logger.Error("failed to rollback tabapay card transaction", "err", err)
			}

			return &providers.TransferResponse{
				Type:                       providers.GMTCARD2ACH,
				OutgoingTransferState:      transactions.StateCompleted,
				OutgoingTransferExternalID: tabapayTransaction.ID,
				IncomingTransferState:      transactions.StateFailed,
				IncomingTransferExternalID: achTransaction.ID,
			}, nil
		}
		return nil, err
	}

	err = workflow.ExecuteActivity(ctx, a.VerifyTransaction, achTransaction.ID).Get(ctx, nil)
	if err != nil {
		logger.Error("failed to verify gmt transaction", "err", err)
		if temporal_utils.IsNonRetryableError(err) {
			err = workflow.ExecuteActivity(newCtx, a.ReverseTabapayTransaction, tabapayTransaction.ID).Get(newCtx, nil)
			if err != nil {
				logger.Error("failed to rollback tabapay card transaction", "err", err)
			}
			return &providers.TransferResponse{
				Type:                       providers.GMTCARD2ACH,
				OutgoingTransferState:      transactions.StateCompleted,
				OutgoingTransferExternalID: tabapayTransaction.ID,
				IncomingTransferState:      transactions.StateFailed,
				IncomingTransferExternalID: achTransaction.ID,
			}, nil
		}
		return nil, err
	}

	var refID string
	err = workflow.ExecuteActivity(ctx, a.CreateWorkflowRef, CreateWorkflowRefArgs{
		ExternalID:    achTransaction.ID,
		WorkflowID:    workflow.GetInfo(ctx).WorkflowExecution.ID,
		WorkflowRunID: workflow.GetInfo(ctx).WorkflowExecution.RunID,
		ActivityName:  "CARD_to_ACH",
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
		if notify.Password != achTransaction.ID {
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

	// update send transfer state.
	err = workflow.ExecuteActivity(ctx, a.UpdateTransferStateByType, recvTrxID, args.ToWalletID, transactions.TransferTypeCreditBankAccount, state).Get(ctx, nil)
	if err != nil {
		logger.Error("failed to update transaction state", "error", err, "state", state)
		return nil, err
	}

	err = workflow.ExecuteActivity(ctx, a.ConfirmPaidNotification, achTransaction.ID).Get(ctx, nil)
	if err != nil {
		logger.Error("failed to clear paid notification", "error", err, "ext ID", achTransaction.ID)
		if temporal_utils.IsNonRetryableError(err) {
			err = workflow.ExecuteActivity(newCtx, a.ReverseTabapayTransaction, tabapayTransaction.ID).Get(newCtx, nil)
			if err != nil {
				logger.Error("failed to rollback tabapay card transaction", "err", err)
			}
			return &providers.TransferResponse{
				Type:                       providers.GMTCARD2ACH,
				OutgoingTransferState:      transactions.StateCompleted,
				OutgoingTransferExternalID: tabapayTransaction.ID,
				IncomingTransferState:      transactions.StateFailed,
				IncomingTransferExternalID: achTransaction.ID,
			}, nil
		}
		return nil, err
	}

	return &providers.TransferResponse{
		Type:                       providers.GMTCARD2ACH,
		OutgoingTransferState:      transactions.StateCompleted,
		OutgoingTransferExternalID: tabapayTransaction.ID,
		IncomingTransferState:      state,
		IncomingTransferExternalID: achTransaction.ID,
	}, nil
}

func ACH2CardTransferWorkflow(ctx workflow.Context, args providers.TransfersArgs) (*providers.TransferResponse, error) {
	var a *Activity

	ao := workflow.ActivityOptions{
		StartToCloseTimeout: 20 * time.Minute,
	}
	ctx = workflow.WithActivityOptions(ctx, ao)

	logger := workflow.GetLogger(ctx)
	logger.Info("ACH2CardTransferWorkflow workflow started", "From", args.FromLinkedAccountID, "Amount", args.Amount)

	err := workflow.ExecuteActivity(ctx, a.CheckWalletOFAC, args.ToWalletID).Get(ctx, nil)
	if err != nil {
		logger.Error("failed to do to linked account OFAC checks", "err", err)
		if temporal_utils.IsNonRetryableError(err) {
			return &providers.TransferResponse{
				Type:                  providers.GMTACH2CARD,
				OutgoingTransferState: transactions.StateFailed,
				IncomingTransferState: transactions.StateFailed,
			}, nil
		}
		return nil, err
	}

	err = workflow.ExecuteActivity(ctx, a.CheckWalletOFAC, args.FromWalletID).Get(ctx, nil)
	if err != nil {
		logger.Error("failed to do from linked account OFAC checks", "err", err)
		if temporal_utils.IsNonRetryableError(err) {
			return &providers.TransferResponse{
				Type:                  providers.GMTACH2CARD,
				OutgoingTransferState: transactions.StateFailed,
				IncomingTransferState: transactions.StateFailed,
			}, nil
		}
		return nil, err
	}

	var cr ComplianceResp
	err = workflow.ExecuteActivity(ctx, a.ACH2CardCompliance, args).Get(ctx, &cr)
	if err != nil {
		logger.Error("failed to do compliance checks", "err", err)
		if temporal_utils.IsNonRetryableError(err) {
			return &providers.TransferResponse{
				Type:                  providers.GMTACH2CARD,
				OutgoingTransferState: transactions.StateFailed,
				IncomingTransferState: transactions.StateFailed,
			}, nil
		}
		return nil, err
	}

	err = workflow.ExecuteActivity(ctx, a.UpdateSendRecvUser, cr).Get(ctx, nil)
	if err != nil {
		logger.Error("failed to upsert gmt send recv user", "err", err)
		return nil, err
	}

	var achTransaction TransactionResp
	err = workflow.ExecuteActivity(ctx, a.InsertACH2Card, args).Get(ctx, &achTransaction)
	if err != nil {
		logger.Error("failed to insert gmt transaction", "err", err)
		if temporal_utils.IsNonRetryableError(err) {
			return &providers.TransferResponse{
				Type:                  providers.GMTACH2CARD,
				OutgoingTransferState: transactions.StateFailed,
				IncomingTransferState: transactions.StateFailed,
			}, nil
		}
		return nil, err
	}

	err = workflow.ExecuteActivity(ctx, a.SaveReceipt, achTransaction).Get(ctx, nil)
	if err != nil {
		logger.Error("failed to save gmt transaction receipt", "err", err)
		return nil, err
	}

	// Insert/update transfers
	err = workflow.ExecuteActivity(ctx, a.AddTransactionTransfer, args.FromTransactionID, []transactions.TransferArgs{
		{
			LinkedAccountID: args.FromLinkedAccountID,
			Type:            transactions.TransferTypeDebitBankAccount,
			Amount:          args.Amount,
			State:           transactions.StatePending,
			ForeignID:       achTransaction.ID,
		},
	}).Get(ctx, nil)
	if err != nil {
		logger.Error("error updating transaction transfer", "Error", err)
		return nil, err
	}

	// TODO: risk Scores if we want

	err = workflow.ExecuteActivity(ctx, a.VerifyTransaction, achTransaction.ID).Get(ctx, nil)
	if err != nil {
		logger.Error("failed to verify gmt transaction", "err", err)
		if temporal_utils.IsNonRetryableError(err) {
			return &providers.TransferResponse{
				Type:                       providers.GMTACH2CARD,
				OutgoingTransferState:      transactions.StateFailed,
				OutgoingTransferExternalID: achTransaction.ID,
				IncomingTransferState:      transactions.StateFailed,
			}, nil
		}
		return nil, err
	}

	var refID string
	err = workflow.ExecuteActivity(ctx, a.CreateWorkflowRef, CreateWorkflowRefArgs{
		ExternalID:    achTransaction.ID,
		WorkflowID:    workflow.GetInfo(ctx).WorkflowExecution.ID,
		WorkflowRunID: workflow.GetInfo(ctx).WorkflowExecution.RunID,
		ActivityName:  "ACH_to_CARD",
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
		if notify.Password != achTransaction.ID {
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

	// update send transfer state
	err = workflow.ExecuteActivity(ctx, a.UpdateTransferStateByType, args.FromTransactionID, args.FromWalletID, transactions.TransferTypeDebitBankAccount, state).Get(ctx, nil)
	if err != nil {
		logger.Error("failed to update transaction state", "error", err, "state", state)
		return nil, err
	}

	err = workflow.ExecuteActivity(ctx, a.ConfirmPaidNotification, achTransaction.ID).Get(ctx, nil)
	if err != nil {
		logger.Error("failed to clear paid notification", "error", err, "ext ID", achTransaction.ID)
		if temporal_utils.IsNonRetryableError(err) {
			return &providers.TransferResponse{
				Type:                       providers.GMTACH2CARD,
				OutgoingTransferState:      transactions.StateFailed,
				OutgoingTransferExternalID: achTransaction.ID,
				IncomingTransferState:      transactions.StateFailed,
			}, nil
		}
		return nil, err
	}

	// Insert incoming transfer
	var recvTrxID string
	err = workflow.SideEffect(ctx, func(ctx workflow.Context) interface{} {
		return uuid.NewString()
	}).Get(&recvTrxID)
	if err != nil {
		logger.Error("error generating transactionID as side effect", "Error", err)
		return &providers.TransferResponse{
			Type:                       providers.GMTACH2CARD,
			OutgoingTransferState:      transactions.StateFailed,
			OutgoingTransferExternalID: achTransaction.ID,
			IncomingTransferState:      transactions.StateFailed,
		}, nil
	}

	var tabapayReferenceID string
	err = workflow.SideEffect(ctx, func(ctx workflow.Context) interface{} {
		return tabapay.NewReferenceID()
	}).Get(&tabapayReferenceID)
	if err != nil {
		logger.Error("error generating tabapay ReferenceID as side effect", "Error", err)
		return &providers.TransferResponse{
			Type:                       providers.GMTACH2CARD,
			OutgoingTransferState:      transactions.StateFailed,
			OutgoingTransferExternalID: achTransaction.ID,
			IncomingTransferState:      transactions.StateFailed,
		}, nil
	}

	newCtx := workflow.WithValue(ctx, httplog.ContextKey, &httplog.Metadata{
		Context: fmt.Sprintf("transactionID=%s", args.FromTransactionID),
	})
	newCtx = workflow.WithActivityOptions(newCtx, workflow.ActivityOptions{
		StartToCloseTimeout: 95 * time.Second, // cmay take up to 90 seconds.,
		RetryPolicy: &temporal.RetryPolicy{
			MaximumAttempts: 2, // tabapay will block us if we keep retrying
		},
	})
	var tabapayTransaction tabapay.Transaction
	err = workflow.ExecuteActivity(newCtx, a.PushToCard, PushToCard{
		TransactionID:       recvTrxID,
		CardLinkedAccountID: args.ToLinkedAccountID,
		Amount:              args.Amount,
		ReferenceID:         tabapayReferenceID,
	}).Get(ctx, &tabapayTransaction)
	if err != nil {
		logger.Error("Failed to push to card.", "Error", err)
		if temporal_utils.IsNonRetryableError(err) {
			// Try to fail tx on GMT
			innerErr := workflow.ExecuteActivity(ctx, a.UpdateCardTransactionStatus, achTransaction.ID, transactions.StateFailed).Get(ctx, nil)
			if innerErr != nil {
				logger.Error("failed to update card transaction on gmt to failed", "err", innerErr)
			}
			return &providers.TransferResponse{
				Type:                       providers.GMTACH2CARD,
				OutgoingTransferState:      state,
				OutgoingTransferExternalID: achTransaction.ID,
				IncomingTransferState:      transactions.StateFailed,
			}, nil
		}
		return nil, err
	}
	if tabapay.IsTransactionStatusUnknown(tabapayTransaction) {
		// check again in 90 sec
		_ = workflow.Sleep(ctx, time.Second*90)
		logger.Info("Tabapay transaction status unknown. Checking again id=", tabapayTransaction.ID)
		err = workflow.ExecuteActivity(newCtx, a.GetTabapayTransaction, tabapayTransaction.ID).Get(newCtx, &tabapayTransaction)
	}
	if err != nil || !tabapay.IsSuccessfulTransaction(tabapayTransaction) {
		logger.Error("Failed to push to card.", "Error", err)
		if temporal_utils.IsNonRetryableError(err) {
			// Try to fail tx on GMT
			innerErr := workflow.ExecuteActivity(ctx, a.UpdateCardTransactionStatus, achTransaction.ID, transactions.StateFailed).Get(ctx, nil)
			if innerErr != nil {
				logger.Error("failed to update card transaction on gmt to failed", "err", innerErr)
			}
			return &providers.TransferResponse{
				Type:                       providers.GMTACH2CARD,
				OutgoingTransferState:      state,
				OutgoingTransferExternalID: achTransaction.ID,
				IncomingTransferState:      transactions.StateFailed,
			}, nil
		}
		return nil, err
	}

	// Notify GMT of completed card transaction.
	err = workflow.ExecuteActivity(ctx, a.UpdateCardTransactionStatus, achTransaction.ID, transactions.StateCompleted).Get(ctx, nil)
	if err != nil {
		logger.Error("failed to update card transaction on gmt", "err", err)
		return nil, err
	}

	// TODO: risk assessment

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
				ForeignID:       achTransaction.ID,
				LinkedAccountID: args.ToLinkedAccountID,
				Type:            transactions.TransferTypeCreditCard,
				Amount:          args.Amount,
				State:           transactions.StateCompleted,
			},
		},
	}).Get(ctx, nil)
	if err != nil {
		logger.Error("failed to add transaction for recipient", "err", err)
		if temporal_utils.IsNonRetryableError(err) {
			return &providers.TransferResponse{
				Type:                       providers.GMTACH2CARD,
				OutgoingTransferState:      state,
				OutgoingTransferExternalID: achTransaction.ID,
				IncomingTransferState:      transactions.StateFailed,
				IncomingTransferExternalID: tabapayTransaction.ID,
			}, nil
		}
		return nil, err
	}

	return &providers.TransferResponse{
		Type:                       providers.GMTACH2CARD,
		OutgoingTransferState:      state,
		OutgoingTransferExternalID: achTransaction.ID,
		IncomingTransferState:      transactions.StateCompleted,
		IncomingTransferExternalID: tabapayTransaction.ID,
	}, nil
}

func Card2CardTransferWorkflow(ctx workflow.Context, args providers.TransfersArgs) (*providers.TransferResponse, error) {
	var a *Activity

	ao := workflow.ActivityOptions{
		StartToCloseTimeout: 20 * time.Minute,
	}
	ctx = workflow.WithActivityOptions(ctx, ao)

	logger := workflow.GetLogger(ctx)
	logger.Info("Card2CardTransferWorkflow workflow started", "From", args.FromLinkedAccountID, "Amount", args.Amount)

	err := workflow.ExecuteActivity(ctx, a.CheckWalletOFAC, args.ToWalletID).Get(ctx, nil)
	if err != nil {
		logger.Error("failed to do to linked account OFAC checks", "err", err)
		if temporal_utils.IsNonRetryableError(err) {
			return &providers.TransferResponse{
				Type:                  providers.GMTCARD2CARD,
				OutgoingTransferState: transactions.StateFailed,
				IncomingTransferState: transactions.StateFailed,
			}, nil
		}
		return nil, err
	}

	err = workflow.ExecuteActivity(ctx, a.CheckWalletOFAC, args.FromWalletID).Get(ctx, nil)
	if err != nil {
		logger.Error("failed to do from linked account OFAC checks", "err", err)
		if temporal_utils.IsNonRetryableError(err) {
			return &providers.TransferResponse{
				Type:                  providers.GMTCARD2CARD,
				OutgoingTransferState: transactions.StateFailed,
				IncomingTransferState: transactions.StateFailed,
			}, nil
		}
		return nil, err
	}

	var cr ComplianceResp
	err = workflow.ExecuteActivity(ctx, a.Card2CardCompliance, args).Get(ctx, &cr)
	if err != nil {
		logger.Error("failed to do compliance checks", "err", err)
		if temporal_utils.IsNonRetryableError(err) {
			return &providers.TransferResponse{
				Type:                  providers.GMTCARD2CARD,
				OutgoingTransferState: transactions.StateFailed,
				IncomingTransferState: transactions.StateFailed,
			}, nil
		}
		return nil, err
	}

	err = workflow.ExecuteActivity(ctx, a.UpdateSendRecvUser, cr).Get(ctx, nil)
	if err != nil {
		logger.Error("failed to upsert gmt send recv user", "err", err)
		return nil, err
	}

	var sendRefID string
	err = workflow.SideEffect(ctx, func(ctx workflow.Context) interface{} {
		return tabapay.NewReferenceID()
	}).Get(&sendRefID)
	if err != nil {
		logger.Error("error generating tabapay send ReferenceID as side effect", "err", err)
		return &providers.TransferResponse{
			Type:                  providers.GMTCARD2CARD,
			OutgoingTransferState: transactions.StateFailed,
			IncomingTransferState: transactions.StateFailed,
		}, nil
	}

	tabapayCtx := workflow.WithValue(ctx, httplog.ContextKey, &httplog.Metadata{
		Context: fmt.Sprintf("transactionID=%s", args.FromTransactionID),
	})
	tabapayCtx = workflow.WithActivityOptions(tabapayCtx, workflow.ActivityOptions{
		StartToCloseTimeout: 95 * time.Second, // cmay take up to 90 seconds.
		RetryPolicy: &temporal.RetryPolicy{
			MaximumAttempts: 2, // tabapay will block us if we keep retrying
		},
	})

	var sendTransaction tabapay.Transaction
	err = workflow.ExecuteActivity(tabapayCtx, a.PullFromCard, PullFromCardArgs{
		TransactionID:       args.FromTransactionID,
		CardLinkedAccountID: args.FromLinkedAccountID,
		Amount:              args.Amount,
		ThreeDSID:           args.ThreeDSID,
		ReferenceID:         sendRefID,
	}).Get(ctx, &sendTransaction)
	if err != nil {
		logger.Error("failed to pull from card", "err", err)
		if temporal_utils.IsNonRetryableError(err) {
			return &providers.TransferResponse{
				Type:                       providers.GMTCARD2CARD,
				OutgoingTransferState:      transactions.StateFailed,
				OutgoingTransferExternalID: sendTransaction.ID,
				IncomingTransferState:      transactions.StateFailed,
			}, nil
		}
		return nil, err
	}
	if tabapay.IsTransactionStatusUnknown(sendTransaction) {
		// check again in 90 sec
		_ = workflow.Sleep(ctx, time.Second*90)
		logger.Info("Tabapay transaction status unknown. Checking again. id=", sendTransaction.ID)
		err = workflow.ExecuteActivity(tabapayCtx, a.GetTabapayTransaction, sendTransaction.ID).Get(tabapayCtx, &sendTransaction)
	}
	if err != nil || !tabapay.IsSuccessfulTransaction(sendTransaction) {
		logger.Error("failed to pull from card", "err", err)
		if temporal_utils.IsNonRetryableError(err) {
			return &providers.TransferResponse{
				Type:                       providers.GMTCARD2CARD,
				OutgoingTransferState:      transactions.StateFailed,
				OutgoingTransferExternalID: sendTransaction.ID,
				IncomingTransferState:      transactions.StateFailed,
			}, nil
		}
		return nil, err
	}

	// Insert/update transfers
	err = workflow.ExecuteActivity(ctx, a.AddTransactionTransfer, args.FromTransactionID, []transactions.TransferArgs{
		{
			LinkedAccountID: args.FromLinkedAccountID,
			Type:            transactions.TransferTypeDebitCard,
			Amount:          args.Amount,
			State:           transactions.StateCompleted,
			ForeignID:       sendTransaction.ID,
		},
	}).Get(ctx, nil)
	if err != nil {
		logger.Error("error updating transaction transfer", "Error", err)
		return nil, err
	}

	// Insert GMT TX
	var gmtTransaction TransactionResp
	err = workflow.ExecuteActivity(ctx, a.InsertCard2Card, args).Get(ctx, &gmtTransaction)
	if err != nil {
		logger.Error("failed to insert gmt transaction", "err", err)
		if temporal_utils.IsNonRetryableError(err) {
			err = workflow.ExecuteActivity(tabapayCtx, a.ReverseTabapayTransaction, sendTransaction.ID).Get(tabapayCtx, nil)
			if err != nil {
				logger.Error("failed to rollback tabapay card transaction", "err", err)
			}
			return &providers.TransferResponse{
				Type:                       providers.GMTCARD2CARD,
				OutgoingTransferState:      transactions.StateCompleted,
				OutgoingTransferExternalID: sendTransaction.ID,
				IncomingTransferState:      transactions.StateFailed,
			}, nil
		}
		return nil, err
	}

	err = workflow.ExecuteActivity(ctx, a.SaveReceipt, gmtTransaction).Get(ctx, nil)
	if err != nil {
		logger.Error("failed to save gmt transaction receipt", "err", err)
		return nil, err
	}

	var recvTrxID string
	err = workflow.SideEffect(ctx, func(ctx workflow.Context) interface{} {
		return uuid.NewString()
	}).Get(&recvTrxID)
	if err != nil {
		logger.Error("error generating recvTrxID as side effect", "Error", err)
		err = workflow.ExecuteActivity(tabapayCtx, a.ReverseTabapayTransaction, sendTransaction.ID).Get(tabapayCtx, nil)
		if err != nil {
			logger.Error("failed to rollback tabapay card transaction", "err", err)
		}
		return &providers.TransferResponse{
			Type:                       providers.GMTCARD2CARD,
			OutgoingTransferState:      transactions.StateCompleted,
			OutgoingTransferExternalID: sendTransaction.ID,
			IncomingTransferState:      transactions.StateFailed,
		}, nil
	}

	var recvRefID string
	err = workflow.SideEffect(ctx, func(ctx workflow.Context) interface{} {
		return tabapay.NewReferenceID()
	}).Get(&recvRefID)
	if err != nil {
		logger.Error("error generating tabapay ReferenceID as side effect", "Error", err)
		err = workflow.ExecuteActivity(tabapayCtx, a.ReverseTabapayTransaction, sendTransaction.ID).Get(tabapayCtx, nil)
		if err != nil {
			logger.Error("failed to rollback tabapay card transaction", "err", err)
		}
		return &providers.TransferResponse{
			Type:                       providers.GMTCARD2CARD,
			OutgoingTransferState:      transactions.StateFailed,
			OutgoingTransferExternalID: sendTransaction.ID,
			IncomingTransferState:      transactions.StateFailed,
		}, nil
	}

	var recvTransaction tabapay.Transaction
	err = workflow.ExecuteActivity(tabapayCtx, a.PushToCard, PushToCard{
		TransactionID:       recvTrxID,
		CardLinkedAccountID: args.ToLinkedAccountID,
		Amount:              args.Amount,
		ReferenceID:         recvRefID,
	}).Get(ctx, &recvTransaction)
	if err != nil {
		logger.Error("Failed to push to card.", "Error", err)
		if temporal_utils.IsNonRetryableError(err) {
			err = workflow.ExecuteActivity(tabapayCtx, a.ReverseTabapayTransaction, sendTransaction.ID).Get(tabapayCtx, nil)
			if err != nil {
				logger.Error("failed to rollback tabapay card transaction", "err", err)
			}
			// Try to fail tx on GMT
			innerErr := workflow.ExecuteActivity(ctx, a.UpdateCardTransactionStatus, gmtTransaction.ID, transactions.StateFailed).Get(ctx, nil)
			if innerErr != nil {
				logger.Error("failed to update card transaction on gmt to failed", "err", innerErr)
			}
			return &providers.TransferResponse{
				Type:                       providers.GMTCARD2CARD,
				OutgoingTransferState:      transactions.StateCompleted,
				OutgoingTransferExternalID: sendTransaction.ID,
				IncomingTransferState:      transactions.StateFailed,
			}, nil
		}
		return nil, err
	}
	if tabapay.IsTransactionStatusUnknown(recvTransaction) {
		// check again in 90 sec
		_ = workflow.Sleep(ctx, time.Second*90)
		logger.Info("Tabapay transaction status unknown. Checking again id=", recvTransaction.ID)
		err = workflow.ExecuteActivity(tabapayCtx, a.GetTabapayTransaction, recvTransaction.ID).Get(tabapayCtx, &recvTransaction)
	}
	if err != nil || !tabapay.IsSuccessfulTransaction(recvTransaction) {
		logger.Error("Failed to push to card.", "Error", err)
		if temporal_utils.IsNonRetryableError(err) {
			err = workflow.ExecuteActivity(tabapayCtx, a.ReverseTabapayTransaction, sendTransaction.ID).Get(tabapayCtx, nil)
			if err != nil {
				logger.Error("failed to rollback tabapay card transaction", "err", err)
			}
			// Try to fail tx on GMT
			innerErr := workflow.ExecuteActivity(ctx, a.UpdateCardTransactionStatus, gmtTransaction.ID, transactions.StateFailed).Get(ctx, nil)
			if innerErr != nil {
				logger.Error("failed to update card transaction on gmt to failed", "err", innerErr)
			}
			return &providers.TransferResponse{
				Type:                       providers.GMTCARD2CARD,
				OutgoingTransferState:      transactions.StateCompleted,
				OutgoingTransferExternalID: sendTransaction.ID,
				IncomingTransferState:      transactions.StateFailed,
			}, nil
		}
		return nil, err
	}

	// Notify GMT of completed card transaction.
	err = workflow.ExecuteActivity(ctx, a.UpdateCardTransactionStatus, gmtTransaction.ID, transactions.StateCompleted).Get(ctx, nil)
	if err != nil {
		logger.Error("failed to update card transaction on gmt", "err", err)
		return nil, err
	}

	// TODO: risk assessment

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
				ForeignID:       recvTransaction.ID,
				LinkedAccountID: args.ToLinkedAccountID,
				Type:            transactions.TransferTypeCreditCard,
				Amount:          args.Amount,
				State:           transactions.StateCompleted,
			},
		},
	}).Get(ctx, nil)
	if err != nil {
		logger.Error("failed to add transaction for recipient", "err", err)
		if temporal_utils.IsNonRetryableError(err) {
			err = workflow.ExecuteActivity(tabapayCtx, a.ReverseTabapayTransaction, sendTransaction.ID).Get(tabapayCtx, nil)
			if err != nil {
				logger.Error("failed to rollback tabapay card transaction", "err", err)
			}
			return &providers.TransferResponse{
				Type:                       providers.GMTCARD2CARD,
				OutgoingTransferState:      transactions.StateCompleted,
				OutgoingTransferExternalID: sendTransaction.ID,
				IncomingTransferState:      transactions.StateFailed,
				IncomingTransferExternalID: recvTransaction.ID,
			}, nil
		}
		return nil, err
	}

	return &providers.TransferResponse{
		Type:                       providers.GMTCARD2CARD,
		OutgoingTransferState:      transactions.StateCompleted,
		OutgoingTransferExternalID: sendTransaction.ID,
		IncomingTransferState:      transactions.StateFailed,
		IncomingTransferExternalID: recvTransaction.ID,
	}, nil
}
