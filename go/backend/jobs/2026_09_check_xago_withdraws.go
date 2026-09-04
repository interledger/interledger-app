package jobs

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/interledger/interledger-app/go/backend/payments"
	"github.com/interledger/interledger-app/go/backend/transactions"
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

type XagoWithdrawTransaction struct {
	TransactionID string `json:"transactionId"`
	WalletID      string `json:"walletId"`
}

func CheckXagoWithdrawsJob(ctx workflow.Context, transactionData XagoWithdrawTransaction) error {

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

	if transaction.Type != transactions.TransactionTypeWithdrawal {
		return fmt.Errorf("not a withdrawal transaction")
	}
	if transaction.State != transactions.StatePending {
		return fmt.Errorf("transaction is not pending")
	}

	var externalTX string
	err := workflow.ExecuteActivity(wfCtx, a.GetXagoTransactionID, transactionData.TransactionID).Get(wfCtx, &externalTX)
	if err != nil {
		return err
	}

	var state string
	if err = workflow.ExecuteActivity(wfCtx, a.SyncWithXagoWithdrawal, externalTX, transactionData.TransactionID).Get(wfCtx, &state); err != nil {
		return err
	}

	if state == "success" || state == "rejected" {
		var transactionStatus = transactions.StateCompleted
		var paymentStatus = payments.StateCompleted
		if state == "rejected" {
			transactionStatus = transactions.StateFailed
			paymentStatus = payments.StateFailed
		}
		if err = workflow.ExecuteActivity(wfCtx, a.UpdateXagoWithdrawTransactionState, transactionData.TransactionID, transactionStatus).Get(wfCtx, nil); err != nil {
			return err
		}
		if err = workflow.ExecuteActivity(wfCtx, a.UpdateXagoWithdrawPaymentState, transaction.ForeignID, paymentStatus).Get(wfCtx, nil); err != nil {
			return err
		}

	} else {
		return fmt.Errorf("withdrawal errored on the xago side: state %q, no action taken on our side", state)
	}

	return nil
}

func (a *Activity) GetXagoTransactionID(ctx context.Context, transactionID string) (string, error) {
	var ids []string
	err := a.b.DB().SelectContext(ctx, &ids, "SELECT id FROM xago_transactions where transaction_id = $1", transactionID)
	if err != nil {
		return "", err
	}

	if len(ids) == 0 {
		return "", fmt.Errorf("no xago transaction found for transaction ID %s", transactionID)
	}

	return ids[0], nil
}

func (a *Activity) GetTransactionByID(ctx context.Context, data XagoWithdrawTransaction) (*transactions.Transaction, error) {
	return a.b.Transactions().GetTransaction(ctx, data.WalletID, data.TransactionID)
}

func (a *Activity) SyncWithXagoWithdrawal(ctx context.Context, externalID, transactionID string) (string, error) {
	wd, err := a.b.Xago().LookupWithdrawal(ctx, externalID)
	if err != nil {
		return "", err
	}

	if strings.EqualFold(wd.Status, "success") {
		if err := a.b.Xago().FinaliseReserve(ctx, transactionID); err != nil {
			return "", err
		}
		return "success", nil
	}
	if strings.EqualFold(wd.Status, "rejected") {
		if err := a.b.Xago().RollbackReserve(ctx, transactionID); err != nil {
			return "", err
		}
		return "rejected", nil
	}
	if strings.EqualFold(wd.Status, "errored") {
		return "", temporal.NewNonRetryableApplicationError("withdrawal from xago errored", "external", nil, "status", wd.Status)
	}

	return wd.Status, nil
}

func (a *Activity) UpdateXagoWithdrawTransactionState(ctx context.Context, transactionID string, state transactions.State) error {
	return a.b.Transactions().SetTransactionState(ctx, transactionID, state)

}

func (a *Activity) UpdateXagoWithdrawPaymentState(ctx context.Context, paymentID string, state payments.State) error {
	_, err := a.b.DB().ExecContext(ctx, "UPDATE payments SET state=$1, updated_at=now() where id=$2", state, paymentID)
	if err != nil {
		return err
	}
	return nil
}
