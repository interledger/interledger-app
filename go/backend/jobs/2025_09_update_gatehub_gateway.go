package jobs

import (
	"context"
	"slices"
	"time"

	"gitlab.com/fynbos/backend/kyc"
	"gitlab.com/fynbos/log"
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"
)

type GatehubWallets struct {
	ExternalID string `db:"external_id"`
	WalletID   string `db:"wallet_id"`
}

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
	var externalUserIDs []GatehubWallets
	err := workflow.ExecuteActivity(ctx, a.GetGatehubUsers).Get(ctx, &externalUserIDs)
	if err != nil {
		return err
	}
	var listOfUnprocessed []string
	err = workflow.ExecuteActivity(ctx, a.LinkUserToGateHubGatewayByExternalID, externalUserIDs).Get(ctx, &listOfUnprocessed)
	if err != nil {
		return err
	}

	err = workflow.ExecuteActivity(ctx, a.UpdateGatehubKYCStatus, externalUserIDs, listOfUnprocessed).Get(ctx, &listOfUnprocessed)
	if err != nil {
		return err
	}

	log.Info("completed SetGatehubGatewayToPaywiserJob", zap.Bool("is Active", params.IsActive), zap.String("region", params.Region), zap.Strings("wallet_ids", params.Wallets), zap.Strings("not changed: ", listOfUnprocessed))
	return nil
}

func (a *Activity) GetGatehubUsers(ctx context.Context, walletID string) ([]GatehubWallets, error) {
	var externalUserIDs []GatehubWallets
	err := a.b.DB().SelectContext(ctx, &externalUserIDs, "SELECT external_id,wallet_id FROM gatehub_users")
	if err != nil {
		return nil, err
	}

	return externalUserIDs, nil
}

func (a *Activity) LinkUserToGateHubGatewayByExternalID(ctx context.Context, wallets []GatehubWallets) ([]string, error) {
	var listOfUnprocessed []string
	for _, externalUser := range wallets {
		if wallets == nil || externalUser.ExternalID == "" {
			continue
		}
		err := a.b.Gatehub().LinkUserToGateHubGatewayByExternalID(ctx, externalUser.ExternalID)
		if err != nil {
			listOfUnprocessed = append(listOfUnprocessed, externalUser.WalletID)
			log.Error("LinkUserToGateHubGatewayByExternalID failed", zap.Any("err", err), zap.String("wallet_id", externalUser.WalletID), zap.Error(err))

		}
	}
	return listOfUnprocessed, nil
}

func (a *Activity) UpdateGatehubKYCStatus(ctx context.Context, wallets []GatehubWallets, listOfUnprocessed []string) ([]string, error) {
	for _, externalUser := range wallets {
		if wallets == nil || externalUser.ExternalID == "" {
			continue
		}
		if slices.Contains(listOfUnprocessed, externalUser.WalletID) {
			continue
		}
		err := a.b.KYC().SetKYCStatus(ctx, externalUser.WalletID, kyc.StatusPending)
		if err != nil {
			listOfUnprocessed = append(listOfUnprocessed, externalUser.WalletID)
			log.Error("UpdateKYCStatus failed", zap.Any("err", err), zap.String("wallet_id", externalUser.WalletID), zap.Error(err))

		}
	}
	return listOfUnprocessed, nil
}
