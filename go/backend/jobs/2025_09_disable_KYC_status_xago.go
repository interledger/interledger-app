package jobs

import (
	"context"
	"strings"
	"time"

	"gitlab.com/fynbos/backend/country"
	"gitlab.com/fynbos/backend/kyc"
	"gitlab.com/fynbos/log"
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"
)

func RestartKYCstatusForXagoJob(ctx workflow.Context, params []string) error {
	var a *Activity
	ao := workflow.ActivityOptions{
		StartToCloseTimeout: 10 * time.Second,
		RetryPolicy: &temporal.RetryPolicy{
			MaximumAttempts: 1,
		},
	}
	ctx = workflow.WithActivityOptions(ctx, ao)
	log.Info("starting RestartKYCstatusForXagoJob", zap.Strings("wallet_ids", params))
	var walletIDs []string = params
	if len(walletIDs) == 0 {
		log.Info("No wallet ids provided, getting all ZA id")
		err := workflow.ExecuteActivity(ctx, a.FetchXagoWalletIds, params).Get(ctx, &walletIDs)
		if err != nil {
			log.Error("FetchXagoWalletIds failed to fetch wallets", zap.Any("err", err))
			return err
		}
	}

	if len(walletIDs) == 0 {
		log.Info("No wallet ids found, exiting job")
		return nil
	}
	err := workflow.ExecuteActivity(ctx, a.UpdatePersonaInquiryXagoStatus, walletIDs).Get(ctx, nil)
	if err != nil {
		log.Error("UpdatePersonaInquiryStatus failed", zap.Any("err", err))
		return err
	}

	err = workflow.ExecuteActivity(ctx, a.RestartXagoKYCForWallets, walletIDs).Get(ctx, nil)
	if err != nil {
		log.Error("RestartKYCForWallets failed", zap.Any("err", err))
		return err
	}
	log.Info("completed RestartKYCstatusForXagoJob", zap.Strings("wallet_ids", params))
	return nil
}

func (a *Activity) UpdatePersonaInquiryXagoStatus(ctx context.Context, walletIds []string) error {
	query := "UPDATE kyc_persona_inquiries SET state = 'expired' where wallet_id in ($1)"
	_, err := a.b.DB().ExecContext(ctx, query, strings.Join(walletIds, ","))
	if err != nil {
		return err
	}

	return nil
}
func (a *Activity) RestartXagoKYCForWallets(ctx context.Context, walletIds []string) error {

	query := "UPDATE wallet_kyc_status SET status = $1 where wallet_id in ($2)"
	_, err := a.b.DB().ExecContext(ctx, query, kyc.StatusUnknown, strings.Join(walletIds, ","))
	if err != nil {
		return err
	}

	return nil
}

func (a *Activity) FetchXagoWalletIds(ctx context.Context) ([]string, error) {
	query := "SELECT id FROM wallets where country=$1"
	var walletList []string

	err := a.b.DB().SelectContext(ctx, &walletList, query, country.ZA)
	if err != nil {
		return nil, err
	}

	return walletList, nil
}
