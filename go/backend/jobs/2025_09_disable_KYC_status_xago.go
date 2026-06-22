package jobs

import (
	"context"
	"time"

	"github.com/interledger/interledger-app/go/backend/kyc"
	"github.com/interledger/interledger-app/go/backend/kyc/persona"
	"github.com/interledger/interledger-app/go/log"
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"
)

func RestartKYCstatusForXagoJob(ctx workflow.Context) error {
	var a *Activity
	ao := workflow.ActivityOptions{
		StartToCloseTimeout: 10 * time.Second,
		RetryPolicy: &temporal.RetryPolicy{
			MaximumAttempts: 1,
		},
	}
	ctx = workflow.WithActivityOptions(ctx, ao)
	err := workflow.ExecuteActivity(ctx, a.UpdatePersonaInquiryXagoStatus).Get(ctx, nil)
	if err != nil {
		log.Error("UpdatePersonaInquiryStatus failed", zap.Any("err", err))
		return err
	}

	err = workflow.ExecuteActivity(ctx, a.RestartXagoKYCForWallets).Get(ctx, nil)
	if err != nil {
		log.Error("RestartKYCForWallets failed", zap.Any("err", err))
		return err
	}
	log.Info("completed RestartKYCstatusForXagoJob")
	return nil
}

func (a *Activity) UpdatePersonaInquiryXagoStatus(ctx context.Context) error {
	query := "UPDATE kyc_persona_inquiries SET state = $1 where wallet_id in (SELECT id FROM wallets where country = 'ZA')"
	_, err := a.b.DB().ExecContext(ctx, query, persona.InquiryExpired)
	if err != nil {
		return err
	}

	return nil
}
func (a *Activity) RestartXagoKYCForWallets(ctx context.Context) error {
	query := "UPDATE wallet_kyc_status SET status = $1 where wallet_id in (SELECT id FROM wallets where country = 'ZA')"
	_, err := a.b.DB().ExecContext(ctx, query, kyc.StatusUnknown)
	if err != nil {
		return err
	}

	return nil
}
