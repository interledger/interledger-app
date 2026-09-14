package jobs

import (
	"context"
	"fmt"
	"time"

	"github.com/interledger/interledger-app/go/backend/currency"
	"github.com/interledger/interledger-app/go/backend/db"
	"github.com/interledger/interledger-app/go/backend/providers/gatehub"
	"github.com/interledger/interledger-app/go/backend/providers/gatehub/external"
	ops_gh "github.com/interledger/interledger-app/go/backend/providers/gatehub/ops"
	"github.com/interledger/interledger-app/go/backend/transactions"
	"github.com/interledger/interledger-app/go/log"
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"
)

type GatehubTransactionDiscrepancy struct {
	TransactionID   string
	ForeignID       string
	Reason          string
	InternalState   transactions.State
	ExternalStatus  string
	InternalType    transactions.TransactionType
	ExternalTransactionType string 
}

type GatehubTransactionsReport struct {
	WalletID        string
	Checked         int
	InternalBalance *currency.Amount
	GatehubBalance  *currency.Amount
	Discrepancies   []GatehubTransactionDiscrepancy
}

func CheckUserGatehubTransactionsJob(ctx workflow.Context, walletID string) (*GatehubTransactionsReport, error) {
	var a *Activity
	wfCtx := workflow.WithActivityOptions(ctx, workflow.ActivityOptions{
		StartToCloseTimeout: 10 * time.Minute,
		RetryPolicy: &temporal.RetryPolicy{
			MaximumAttempts: 3,
		},
	})

	var report *GatehubTransactionsReport
	if err := workflow.ExecuteActivity(wfCtx, a.CheckUserGatehubTransactions, walletID).Get(wfCtx, &report); err != nil {
		return nil, err
	}

	return report, nil
}

func (a *Activity) CheckUserGatehubTransactions(ctx context.Context, walletID string) (*GatehubTransactionsReport, error) {
	u, err := a.b.Gatehub().GetUser(ctx, walletID)
	if err != nil {
		return nil, fmt.Errorf("getting gatehub user for wallet %s: %w", walletID, err)
	}

	externalTxs, err := a.b.Gatehub().ExternalClient().GetUserTransactions(ctx, u.UUID)
	if err != nil {
		return nil, fmt.Errorf("listing gatehub transactions for wallet %s: %w", walletID, err)
	}

	internalTxs, err := listWalletGatehubTransactions(ctx, a.b, walletID)
	if err != nil {
		return nil, fmt.Errorf("listing internal transactions for wallet %s: %w", walletID, err)
	}

	internalByForeignID := make(map[string]transactions.Transaction, len(internalTxs))
	for _, t := range internalTxs {
		internalByForeignID[t.ForeignID] = t
	}

	report := &GatehubTransactionsReport{WalletID: walletID}

	internalBalance, gatehubBalance, err := gatehubBalances(ctx, a.b, walletID, u.UUID)
	if err != nil {
		return nil, fmt.Errorf("getting balances for wallet %s: %w", walletID, err)
	}
	report.InternalBalance = internalBalance
	report.GatehubBalance = gatehubBalance

	matched := make(map[string]bool, len(externalTxs))

	for _, ext := range externalTxs {
		report.Checked++

		internal, ok := internalByForeignID[ext.ID]
		if !ok {
			report.Discrepancies = append(report.Discrepancies, GatehubTransactionDiscrepancy{
				ForeignID:       ext.ID,
				Reason:          "gatehub transaction has no matching internal transaction",
				InternalState:   transactionStateNotFound,
				ExternalStatus:  gatehubStatusName(ext.Status),
				InternalType:    transactionTypeNotFound,
				ExternalTransactionType: gatehubTransactionTypeName(ext.Type),
			})
			continue
		}
		matched[ext.ID] = true

		if wantState := gatehubStatusToState(ext.Status); wantState != "" && internal.State != wantState {
			report.Discrepancies = append(report.Discrepancies, GatehubTransactionDiscrepancy{
				TransactionID:   internal.ID,
				ForeignID:       ext.ID,
				Reason:          fmt.Sprintf("state mismatch: internal=%s external_status=%s", internal.State, gatehubStatusName(ext.Status)),
				InternalState:   internal.State,
				ExternalStatus:  gatehubStatusName(ext.Status),
				InternalType:    internal.Type,
				ExternalTransactionType: gatehubTransactionTypeName(ext.Type),
			})
		}

		if externalAmount, err := ops_gh.StringToScaledUInt(ext.Total); err == nil && externalAmount != internal.Amount.Value {
			report.Discrepancies = append(report.Discrepancies, GatehubTransactionDiscrepancy{
				TransactionID:   internal.ID,
				ForeignID:       ext.ID,
				Reason:          fmt.Sprintf("amount mismatch: internal=%d external=%d", internal.Amount.Value, externalAmount),
				InternalState:   internal.State,
				ExternalStatus:  gatehubStatusName(ext.Status),
				InternalType:    internal.Type,
				ExternalTransactionType: gatehubTransactionTypeName(ext.Type),
			})
		}
	}

	for foreignID, internal := range internalByForeignID {
		if matched[foreignID] {
			continue
		}
		report.Discrepancies = append(report.Discrepancies, GatehubTransactionDiscrepancy{
			TransactionID:   internal.ID,
			ForeignID:       foreignID,
			Reason:          "internal transaction has no matching gatehub transaction",
			InternalState:   internal.State,
			ExternalStatus:  gatehubStatusNotFound,
			InternalType:    internal.Type,
			ExternalTransactionType: gatehubStatusNotFound,
		})
	}

	if len(report.Discrepancies) > 0 {
		log.Info("gatehub transaction discrepancies found",
			zap.String("wallet_id", walletID),
			zap.Int("count", len(report.Discrepancies)))
	}

	return report, nil
}

func gatehubBalances(ctx context.Context, b Backends, walletID, externalUserID string) (internal, gatehubBal *currency.Amount, err error) {
	linkedAccounts, err := b.LinkedAccounts().ListByWalletId(ctx, walletID)
	if err != nil {
		return nil, nil, fmt.Errorf("listing linked accounts for wallet %s: %w", walletID, err)
	}

	var laID, laProviderID string
	var laCurrency currency.Currency
	found := false
	for _, la := range linkedAccounts {
		if la.Provider == gatehub.ProviderName && la.Type == gatehub.AccTypeBalance && la.DeletedAt.Time.IsZero() {
			laID, laProviderID, laCurrency = la.ID, la.ProviderID, la.SendCurrency
			found = true
			break
		}
	}
	if !found {
		return nil, nil, nil
	}

	bal, err := b.Gatehub().GetBalance(ctx, laID)
	if err != nil {
		return nil, nil, fmt.Errorf("getting internal balance for linked account %s: %w", laID, err)
	}
	internal = &bal.Total

	extBalances, err := b.Gatehub().ExternalClient().GetWalletBalances(ctx, externalUserID, laProviderID)
	if err != nil {
		return internal, nil, fmt.Errorf("getting gatehub balance for linked account %s: %w", laProviderID, err)
	}
	if len(extBalances) > 0 {
		value, err := ops_gh.StringToScaledUInt(extBalances[0].Total)
		if err != nil {
			return internal, nil, fmt.Errorf("parsing gatehub balance for linked account %s: %w", laProviderID, err)
		}
		amount := currency.FromUInt64(value, laCurrency)
		gatehubBal = &amount
	}

	return internal, gatehubBal, nil
}

func listWalletGatehubTransactions(ctx context.Context, b Backends, walletID string) ([]transactions.Transaction, error) {
	const pageSize = 50

	var result []transactions.Transaction
	page := db.Pagination{PageSize: pageSize}

	for {
		txs, err := b.Transactions().List(ctx, page, walletID)
		if err != nil {
			return nil, err
		}

		for i, t := range txs {
			if i == page.PageSize {
				page.PageToken = t.ID
				break
			}
			if t.Provider == gatehub.ProviderName && t.ForeignID != "" {
				result = append(result, t)
			}
		}

		if len(txs) <= page.PageSize {
			break
		}
	}

	return result, nil
}

func gatehubStatusToState(status int) transactions.State {
	switch status {
	case external.TransactionStatusCompleted:
		return transactions.StateCompleted
	case external.TransactionStatusFailed, external.TransactionStatusUserCancelled, external.TransactionStatusAdminCancelled:
		return transactions.StateFailed
	case external.TransactionStatusPending, external.TransactionStatusProcessing, external.TransactionStatusUnmatched, external.TransactionStatusReturning, external.TransactionStatusManualReview:
		return transactions.StatePending
	default:
		return ""
	}
}
