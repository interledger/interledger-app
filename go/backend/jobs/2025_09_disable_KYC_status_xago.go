package jobs

import (
	"context"
	"strings"
	"time"

	"gitlab.com/fynbos/log"
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"
)

type KYCRestartWallets struct {
	Wallets []string `json:"wallets"`
}

func RestartKYCstatusForXagoJob(ctx workflow.Context, params KYCRestartWallets) error {
	var a *Activity
	ao := workflow.ActivityOptions{
		StartToCloseTimeout: 10 * time.Second,
		RetryPolicy: &temporal.RetryPolicy{
			MaximumAttempts: 1,
		},
	}
	ctx = workflow.WithActivityOptions(ctx, ao)
	log.Info("starting RestartKYCstatusForXagoJob", zap.Strings("wallet_ids", params.Wallets))
	var walletIDs []string = params.Wallets
	if len(walletIDs) == 0 {
		log.Info("No wallet ids provided, getting all ZA id")
		err := workflow.ExecuteActivity(ctx, a.FetchWalletIds, params).Get(ctx, &walletIDs)
		if err != nil {
			log.Error("RestartKYCstatusForXagoJob failed to fetch wallets", zap.Any("err", err))
			return err
		}
	}

	if walletIDs == nil || len(walletIDs) == 0 {
		log.Info("No wallet ids found, exiting job")
		return nil
	}
	err := workflow.ExecuteActivity(ctx, a.UpdatePersonaInquiryStatus, walletIDs).Get(ctx, nil)
	if err != nil {
		log.Error("UpdatePersonaInquiryStatus failed", zap.Any("err", err))
		return err
	}

	err = workflow.ExecuteActivity(ctx, a.RestartKYCForWallets, walletIDs).Get(ctx, nil)
	if err != nil {
		log.Error("RestartKYCForWallets failed", zap.Any("err", err))
		return err
	}
	log.Info("completed RestartKYCstatusForXagoJob", zap.Strings("wallet_ids", params.Wallets))
	return nil
}

func (a *Activity) UpdatePersonaInquiryStatus(ctx context.Context, walletIds []string) error {
	query := "UPDATE  kyc_persona_inquiries SET state = 'expired' where wallet_id in ($1)"
	var walletList []string

	err := a.b.DB().SelectContext(ctx, &walletList, query, strings.Join(walletIds, ","))
	if err != nil {
		return err
	}

	return nil
}
func (a *Activity) RestartKYCForWallets(ctx context.Context, walletIds []string) error {
	query := "UPDATE  wallet_kyc_status SET status = 0 where wallet_id in ($1)"
	var walletList []string

	err := a.b.DB().SelectContext(ctx, &walletList, query, strings.Join(walletIds, ","))
	if err != nil {
		return err
	}

	return nil
}

func (a *Activity) FetchWalletIds(ctx context.Context) ([]string, error) {
	query := "SELECT wallet_id FROM wallets where country='ZA'"
	var walletList []string

	err := a.b.DB().SelectContext(ctx, &walletList, query)
	if err != nil {
		return nil, err
	}

	return walletList, nil
}
