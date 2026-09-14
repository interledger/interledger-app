package jobs

import (
	"context"
	"fmt"
	"time"

	"github.com/interledger/interledger-app/go/backend/payments"
	gatehub_external "github.com/interledger/interledger-app/go/backend/providers/gatehub/external"
	"github.com/interledger/interledger-app/go/backend/transactions"
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

func FixGateHubTransactionJob(ctx workflow.Context, transactionData CheckTransaction) error {

	var a *Activity
	wfCtx := workflow.WithActivityOptions(ctx, workflow.ActivityOptions{
		StartToCloseTimeout: 10 * time.Minute,
		RetryPolicy: &temporal.RetryPolicy{
			MaximumAttempts: 3,
		},
	})

	var transaction *transactions.Transaction
	if err := workflow.ExecuteActivity(wfCtx, a.GetTransactionByID, transactionData).Get(wfCtx, &transaction); err != nil {
		return err
	}

	if transaction.Provider != "gatehub" {
		return fmt.Errorf("not a gatehub transaction")
	}

	if transaction.State == transactions.StateCompleted || transaction.State == transactions.StateFailed {
		return fmt.Errorf("transaction is in final state")
	}

	var state string
	if err := workflow.ExecuteActivity(wfCtx, a.GetGateHubTransactionStatus, transaction.WalletID, transaction.ForeignID).Get(wfCtx, &state); err != nil {
		return err
	}

	switch state {
	case "success", "rejected":
		transactionStatus := transactions.StateCompleted
		paymentStatus := payments.StateCompleted
		if state == "rejected" {
			transactionStatus = transactions.StateFailed
			paymentStatus = payments.StateFailed
		}

		if err := workflow.ExecuteActivity(wfCtx, a.UpdateTransactionStateJob, transactionData.TransactionID, transactionStatus).Get(wfCtx, nil); err != nil {
			return err
		}
		if err := workflow.ExecuteActivity(wfCtx, a.UpdatePaymentStateJob, transaction.ForeignID, paymentStatus).Get(wfCtx, nil); err != nil {
			return err
		}
	default:
		return fmt.Errorf("gatehub transaction %s not yet final: state %q, no action taken", transactionData.TransactionID, state)
	}

	return nil
}

func (a *Activity) GetGateHubTransactionStatus(ctx context.Context, walletID, transactionID string) (string, error) {
	u, err := a.b.Gatehub().GetUser(ctx, walletID)
	if err != nil {
		return "", err
	}

	wd, err := a.b.Gatehub().ExternalClient().GetTransaction(ctx, u.UUID, transactionID)
	if err != nil {
		return "", err
	}

	if wd.Status == gatehub_external.TransactionStatusCompleted {
		if err := a.b.Gatehub().FinaliseReserve(ctx, transactionID); err != nil {
			return "", err
		}
		return "success", nil
	}
	if wd.Status == gatehub_external.TransactionStatusFailed ||
		wd.Status == gatehub_external.TransactionStatusUserCancelled ||
		wd.Status == gatehub_external.TransactionStatusAdminCancelled {
		if err := a.b.Gatehub().RollbackReserve(ctx, transactionID); err != nil {
			return "", err
		}
		return "rejected", nil
	}

	return gatehubStatusName(wd.Status), nil

}
