package jobs

import (
	"context"
	"slices"

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

func SetGatehubGatewayToPaywiserJob(ctx workflow.Context) error {
	var a *Activity
	ao := workflow.ActivityOptions{
		RetryPolicy: &temporal.RetryPolicy{
			MaximumAttempts: 1,
		},
	}
	ctx = workflow.WithActivityOptions(ctx, ao)
	log.Info("starting SetGatehubGatewayToPaywiserJob")
	var gatehubWallets []GatehubWallets
	err := workflow.ExecuteActivity(ctx, a.GetGatehubUsers).Get(ctx, &gatehubWallets)
	if err != nil {
		return err
	}
	var unprocessedGatewayWallets []string
	err = workflow.ExecuteActivity(ctx, a.LinkUserToGatewayByExternalID, gatehubWallets).Get(ctx, &unprocessedGatewayWallets)
	if err != nil {
		return err
	}
	var unprocessedKYCWallets []string
	err = workflow.ExecuteActivity(ctx, a.UpdateGatehubKYCStatus, gatehubWallets, unprocessedGatewayWallets).Get(ctx, &unprocessedKYCWallets)
	if err != nil {
		return err
	}

	log.Info("completed SetGatehubGatewayToPaywiserJob", zap.Strings("wallets with unchanged Gateway ", unprocessedGatewayWallets), zap.Strings("wallets with unchanged KYC status ", unprocessedKYCWallets))
	return nil
}

func (a *Activity) GetGatehubUsers(ctx context.Context) ([]GatehubWallets, error) {
	var gatehubWallets []GatehubWallets
	err := a.b.DB().SelectContext(ctx, &gatehubWallets, "SELECT external_id, wallet_id FROM gatehub_users")
	if err != nil {
		return nil, err
	}

	return gatehubWallets, nil
}

func (a *Activity) LinkUserToGatewayByExternalID(ctx context.Context, wallets []GatehubWallets) ([]string, error) {
	var listOfUnprocessed []string
	for _, externalUser := range wallets {
		if wallets == nil || externalUser.ExternalID == "" {
			continue
		}

		err := a.b.Gatehub().LinkUserToGatewayByExternalID(ctx, externalUser.ExternalID)
		if err != nil {
			log.Error("error linking user to gateway by external ID", zap.String("external_id", externalUser.ExternalID), zap.String("wallet_id", externalUser.WalletID), zap.Error(err))
			listOfUnprocessed = append(listOfUnprocessed, externalUser.WalletID)
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
		}
	}
	return listOfUnprocessed, nil
}
