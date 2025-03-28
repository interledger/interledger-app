package jobs

import (
	"context"
	"time"

	"gitlab.com/fynbos/log"
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

func AddManageCardsFeatureFlagJob(ctx workflow.Context) error {
	var a *Activity
	ao := workflow.ActivityOptions{
		StartToCloseTimeout: 10 * time.Second,
		RetryPolicy: &temporal.RetryPolicy{
			MaximumAttempts: 3,
		},
	}
	ctx = workflow.WithActivityOptions(ctx, ao)
	log.Info("Starting job AddManageCardsFeatureFlagJob")

	err := workflow.ExecuteActivity(ctx, a.UpdateWalletWalletFeaturesActivity).Get(ctx, nil)
	if err != nil {
		return err
	}
	log.Info("Completed job AddManageCardsFeatureFlagJob")
	return nil
}

func (a *Activity) UpdateWalletWalletFeaturesActivity(ctx context.Context) error {
	_, err := a.b.DB().ExecContext(ctx, "ALTER TABLE \"wallet_features\" ADD COLUMN \"manage_cards_enabled\" bool NOT NULL DEFAULT 'false'")

	return err
}
