package jobs

import (
	"context"
	"fmt"

	"github.com/interledger/interledger-app/go/backend/payments"
	"github.com/interledger/interledger-app/go/backend/providers/gatehub/external"
	"github.com/interledger/interledger-app/go/backend/transactions"
)

func (a *Activity) UpdateTransactionStateJob(ctx context.Context, transactionID string, state transactions.State) error {
	return a.b.Transactions().SetTransactionState(ctx, transactionID, state)

}

func (a *Activity) UpdatePaymentStateJob(ctx context.Context, paymentID string, state payments.State) error {
	_, err := a.b.DB().ExecContext(ctx, "UPDATE payments SET state=$1, updated_at=now() where id=$2", state, paymentID)
	if err != nil {
		return err
	}
	return nil
}

func (a *Activity) GetTransactionByID(ctx context.Context, data CheckTransaction) (*transactions.Transaction, error) {
	return a.b.Transactions().GetTransaction(ctx, data.WalletID, data.TransactionID)
}

const (
	transactionStateNotFound = transactions.State("not_found")
	transactionTypeNotFound  = transactions.TransactionType("not_found")
	gatehubStatusNotFound    = "not_found"
)

func gatehubStatusName(status int) string {
	switch status {
	case external.TransactionStatusPending:
		return "pending"
	case external.TransactionStatusProcessing:
		return "processing"
	case external.TransactionStatusUnmatched:
		return "unmatched"
	case external.TransactionStatusReturning:
		return "returning"
	case external.TransactionStatusManualReview:
		return "manual_review"
	case external.TransactionStatusCompleted:
		return "completed"
	case external.TransactionStatusFailed:
		return "failed"
	case external.TransactionStatusUserCancelled:
		return "user_cancelled"
	case external.TransactionStatusAdminCancelled:
		return "admin_cancelled"
	default:
		return fmt.Sprintf("unknown(%d)", status)
	}
}

func gatehubTransactionTypeName(txType int) string {
	switch txType {
	case external.TransactionTypeWithdrawal:
		return "withdrawal"
	case external.TransactionTypeDeposit:
		return "deposit"
	case external.TransactionTypeHosted:
		return "hosted"
	case external.TransactionTypeExchange:
		return "exchange"
	default:
		return fmt.Sprintf("unknown(%d)", txType)
	}
}
