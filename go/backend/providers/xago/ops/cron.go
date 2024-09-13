package ops

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"gitlab.com/fynbos/backend/db"

	"gitlab.com/fynbos/backend/currency"
	"gitlab.com/fynbos/backend/linkedaccounts"
	"gitlab.com/fynbos/backend/providers/xago"
	"gitlab.com/fynbos/backend/providers/xago/external"
	"gitlab.com/fynbos/backend/transactions"
	"gitlab.com/fynbos/log"
	"gitlab.com/fynbos/pacioli"
	"go.temporal.io/api/enums/v1"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"
)

func StartDepositsPolling(b ActivityBackends) {
	// This workflow ID can be user business logic identifier as well.
	workflowID := "cron_xago_deposits_poll"
	workflowOptions := client.StartWorkflowOptions{
		ID:                    workflowID,
		TaskQueue:             "backend",
		CronSchedule:          "0 */1 * * *",                                       // Every hour
		WorkflowIDReusePolicy: enums.WORKFLOW_ID_REUSE_POLICY_TERMINATE_IF_RUNNING, // There can be only one
	}

	we, err := b.Temporal().ExecuteWorkflow(context.Background(), workflowOptions, XagoDepositPollWorkflow)
	if err != nil {
		log.Fatal("Unable to execute workflow", zap.Error(err))
	}
	log.Info("Started workflow", zap.String("WorkflowID", we.GetID()), zap.String("RunID", we.GetRunID()))
}

func XagoDepositPollWorkflow(ctx workflow.Context) error {
	var a *Activity
	ao := workflow.ActivityOptions{
		StartToCloseTimeout: 20 * time.Minute,
	}
	ctx = workflow.WithActivityOptions(ctx, ao)

	logger := workflow.GetLogger(ctx)

	logger.Info("starting xago deposits")

	var deposits []external.Deposit
	err := workflow.ExecuteActivity(ctx, a.PollDeposits).Get(ctx, &deposits)
	if err != nil {
		return err
	}
	// No new deposits, so nothing to do
	if len(deposits) == 0 {
		return nil
	}

	logger.Info("Adding new deposits", "len", len(deposits))

	err = workflow.ExecuteActivity(ctx, a.CreateDepositTransactions, deposits).Get(ctx, nil)
	if err != nil {
		return err
	}

	err = workflow.ExecuteActivity(ctx, a.SaveDeposits, deposits).Get(ctx, nil)
	if err != nil {
		return err
	}

	return nil
}

func (a *Activity) PollDeposits(ctx context.Context) ([]external.Deposit, error) {
	var page int = 1
	var deposits []external.Deposit
	for {
		deps, err := a.b.External().ListDeposits(ctx, page)
		if err != nil {
			return nil, err
		}

		var lastDepID string
		for _, dep := range deps {
			if !strings.EqualFold(dep.Status, "success") {
				continue
			}
			lastDepID = dep.TransactionID
			deposits = append(deposits, dep)
		}

		// Lookup if we already have the deposit in our DB
		if lastDepID != "" {
			var txID string
			err = a.b.DB().GetContext(ctx, &txID, "SELECT transaction_id FROM xago_deposits WHERE transaction_id=$1", lastDepID)
			if err != nil && !errors.Is(err, sql.ErrNoRows) {
				return nil, err
			}

			/// We already have this deposit in our DB so we've reached to where we want to scan
			if txID != "" {
				break
			}
		}
		if len(deps) < 10 {
			break
		}
		page++
	}

	return deposits, nil
}

func (a *Activity) SaveDeposits(ctx context.Context, deposits []external.Deposit) error {
	stmt, err := a.b.DB().PrepareContext(ctx, "INSERT INTO xago_deposits (transaction_id, origin_amount, amount, status, account_id) VALUES ($1, $2,$3, $4,$5)")
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, dep := range deposits {
		_, err = stmt.ExecContext(ctx, dep.TransactionID, dep.OriginAmount, dep.Amount, dep.Status, dep.AccountID)
		if db.IsErrorCode(err, db.UniqueViolationError) {
			continue
		}
		if err != nil {
			return err
		}

		subAcc, err := LookupByAccountID(ctx, a.b, dep.AccountID)
		if err != nil {
			return err
		}

		// Best effort
		a.b.Email().SendDepositReceivedEmail(ctx, subAcc.WalletID, currency.FromFloat64(dep.Amount, currency.ZAR), "", "")
	}

	return nil
}

func (a *Activity) CreateDepositTransactions(ctx context.Context, deposits []external.Deposit) error {
	for _, dep := range deposits {
		subAcc, err := LookupByAccountID(ctx, a.b, dep.AccountID)
		if err != nil {
			return err
		}

		lal, err := a.b.LinkedAccounts().ListByWalletId(ctx, subAcc.WalletID)
		if err != nil {
			return err
		}

		var acc linkedaccounts.LinkedAccount
		for _, la := range lal {
			if la.Provider == xago.ProviderName && la.Type == xago.AccTypeBalance && la.SendCurrency == currency.ZAR {
				acc = la
				break
			}
		}
		if acc.ID == "" {
			return fmt.Errorf("%w no ZAR account found for depost", xago.ErrNotFound)
		}

		// Som of these may actually be no ops because the were filled in by the webhook. All of it's idempotent so it's safe to rerun
		// Idempotent call
		_, err = a.b.Transactions().GetTransaction(ctx, acc.WalletID, dep.TransactionID)
		if errors.Is(err, transactions.ErrNotFound) {
			_, err = a.b.Transactions().CreateTransaction(ctx, transactions.CreateTransactionArgs{
				ID:                      dep.TransactionID,
				WalletID:                acc.WalletID,
				ForeignID:               dep.TransactionID,
				ForeignType:             transactions.TransactionTypeDeposit,
				Provider:                transactions.ProviderXago,
				State:                   transactions.StateCompleted,
				Note:                    "Deposit received",
				Source:                  "Bank Deposit",
				Destination:             acc.WalletID,
				Amount:                  currency.FromFloat64(dep.Amount, currency.ZAR),
				LinkedAccountTitle:      acc.Title(),
				DestinationIdentity:     acc.WalletID,
				DestinationIdentityType: "WalletID",
			})
		}
		if err != nil {
			return err
		}

		// Also an idempotent call
		tr, err := a.b.Pacioli().CreateTransfers(ctx, []pacioli.CreateTransferArgs{
			{
				ID:              dep.TransactionID,
				Amount:          currency.FromFloat64(dep.Amount, currency.ZAR).Value,
				DebitAccountID:  xago.ZAROpsAccount,
				CreditAccountID: acc.ID,
				Pending:         false,
				Code:            1,
				Ledger:          xago.LedgerIDZAR,
			},
		})
		if err != nil {
			return err
		}

		if len(tr) > 0 && tr[0].Code != 0 {
			return fmt.Errorf("failed to create pacioli transaction for xago deposit status (%s)", tr[0].Code)
		}
	}

	return nil
}
