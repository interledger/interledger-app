package jobs

import (
	"context"
	"time"

	"gitlab.com/fynbos/log"
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"
)

func SetGatehubGatewayToPaywiserJob(ctx workflow.Context, params WalletActive) error {
	var a *Activity
	ao := workflow.ActivityOptions{
		StartToCloseTimeout: 10 * time.Second,
		RetryPolicy: &temporal.RetryPolicy{
			MaximumAttempts: 1,
		},
	}
	ctx = workflow.WithActivityOptions(ctx, ao)
	log.Info("starting SetGatehubGatewayToPaywiserJob", zap.Bool("is Active", params.IsActive), zap.String("region", params.Region), zap.Strings("wallet_ids", params.Wallets))
	var externalUserIDs []string
	err := workflow.ExecuteActivity(ctx, a.GetGatehubUsers).Get(ctx, &externalUserIDs)
	if err != nil {
		return err
	}
	var listOfUnprocessed []string
	err = workflow.ExecuteActivity(ctx, a.LinkUserToGateHubGatewayByExternalID, externalUserIDs).Get(ctx, &listOfUnprocessed)
	if err != nil {
		return err
	}

	log.Info("completed SetGatehubGatewayToPaywiserJob", zap.Bool("is Active", params.IsActive), zap.String("region", params.Region), zap.Strings("wallet_ids", params.Wallets), zap.Strings("not changed: ", listOfUnprocessed))
	return nil
}

func (a *Activity) GetGatehubUsers(ctx context.Context, walletID string) ([]string, error) {
	var externalUserIDs []string
	err := a.b.DB().SelectContext(ctx, &externalUserIDs, "SELECT external_id FROM gatehub_users")
	if err != nil {
		return nil, err
	}

	return externalUserIDs, nil
}

func (a *Activity) LinkUserToGateHubGatewayByExternalID(ctx context.Context, externalUserIDs []string) ([]string, error) {
	var listOfUnprocessed []string
	for _, externalUser := range externalUserIDs {
		if externalUserIDs == nil || externalUser == "" {
			continue
		}
		err := a.b.Gatehub().LinkUserToGateHubGatewayByExternalID(ctx, externalUser)
		if err != nil {
			listOfUnprocessed = append(listOfUnprocessed, externalUser)
			log.Error("LinkUserToGateHubGatewayByExternalID failed", zap.Any("err", err), zap.String("external_id", externalUser), zap.Error(err))

		}
	}
	return listOfUnprocessed, nil
}
