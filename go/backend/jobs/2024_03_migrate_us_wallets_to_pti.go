package jobs

import (
	"context"
	"time"

	"github.com/interledger/interledger-app/go/backend/country"
	"github.com/interledger/interledger-app/go/backend/currency"
	"github.com/interledger/interledger-app/go/backend/providers/pti"
	pti_ops "github.com/interledger/interledger-app/go/backend/providers/pti/ops"
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

func MigrateUSWalletsToPTIJob(ctx workflow.Context) ([]string, error) {
	var a *Activity
	ao := workflow.ActivityOptions{
		StartToCloseTimeout: 10 * time.Second,
		RetryPolicy: &temporal.RetryPolicy{
			MaximumAttempts: 3,
		},
	}
	ctx = workflow.WithActivityOptions(ctx, ao)

	var usWalletIDs []string
	err := workflow.ExecuteActivity(ctx, a.ListUSWallets).Get(ctx, &usWalletIDs)
	if err != nil {
		return nil, err
	}

	var results []workflow.Future
	for _, w := range usWalletIDs {
		future := workflow.ExecuteChildWorkflow(ctx, pti_ops.CreateWalletWorkflow, pti.CreateWalletArgs{
			WalletID: w,
			Currency: currency.USD,
			Nickname: "USD Balance",
			Title:    "USD Balance",
		})
		results = append(results, future)
	}

	var failedWalletIDs []string
	for i, result := range results {
		err := result.Get(ctx, nil)
		if err != nil {
			failedWalletIDs = append(failedWalletIDs, usWalletIDs[i])
		}
	}

	err = workflow.ExecuteActivity(ctx, a.SoftDeleteTabapayCards).Get(ctx, nil)
	if err != nil {
		return failedWalletIDs, err
	}

	return failedWalletIDs, nil
}

func (a *Activity) ListUSWallets(ctx context.Context) ([]string, error) {
	var ret []string

	err := a.b.DB().SelectContext(ctx, &ret, "SELECT id from wallets WHERE country=$1;", country.US.String())
	if err != nil {
		return nil, err
	}

	return ret, nil
}

func (a *Activity) SoftDeleteTabapayCards(ctx context.Context) error {
	now := time.Now()
	_, err := a.b.DB().ExecContext(ctx, "UPDATE linked_accounts SET deleted_at=$1, updated_at=$2 WHERE provider=$3;", now, now, "tabapay")
	if err != nil {
		return err
	}

	return nil
}
