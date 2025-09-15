package jobs

import (
	"context"
	"fmt"
	"time"

	"gitlab.com/fynbos/backend/currency"
	"gitlab.com/fynbos/backend/providers/gatehub"
	"gitlab.com/fynbos/backend/transactions"
	"gitlab.com/fynbos/log"
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"
)

func CoverFeesJob(ctx workflow.Context) error {
	var a *Activity
	ao := workflow.ActivityOptions{
		StartToCloseTimeout: 10 * time.Second,
		RetryPolicy: &temporal.RetryPolicy{
			MaximumAttempts: 1,
		},
	}
	ctx = workflow.WithActivityOptions(ctx, ao)

	err := workflow.ExecuteActivity(ctx, a.CoverFees).Get(ctx, nil)
	if err != nil {
		return err
	}

	return nil
}

func (a *Activity) CoverFees(ctx context.Context) error {

	var txs []transactions.DbTransaction
	fields := "t.id,t.foreign_id,t.wallet_id,t.reference_id,t.type,t.state,t.provider,t.source,t.destination,t.title,t.amount,t.provider_fee,t.asset_scale,t.asset_code,t.updated_at,t.linked_account_title,t.destination_identity_type,t.destination_identity,t.reference,t.refund_state"
	err := a.b.DB().SelectContext(ctx, &txs,
		fmt.Sprintf(`SELECT
    		%s
		FROM
			transactions t
			LEFT JOIN transactions r ON t.reference_id = r.id
		WHERE
			t.state = 'Completed'
			AND t.provider_fee > 0
			AND t.type = 'deposit'
			AND r.id IS NULL;
		`, fields))
	if err != nil {
		return err
	}

	for _, tx := range txs {
		log.Debug("Processing transaction:", zap.String("tx", tx.ID))

		var providerID string
		lal, err := a.b.LinkedAccounts().ListByWalletId(ctx, tx.WalletID.String)
		if err != nil {
			fmt.Println("Error la:", err)
			return err
		}
		for _, la := range lal {
			if la.Provider == gatehub.ProviderName && la.Type == gatehub.AccTypeBalance && la.DeletedAt.Time.IsZero() {
				providerID = la.ProviderID
			}
		}
		log.Debug("Found provider ID:", zap.String("provider", providerID))

		if tx.Provider == gatehub.ProviderName {

			args := gatehub.CoverFeeArgs{
				TransactionID: tx.ID,
				ProviderID:    providerID,
				Amount:        currency.FromUInt64(tx.ProviderFee, currency.EUR),
			}
			log.Debug("Args:", zap.Any("arg", args))
			_, err := a.b.Gatehub().RefundProviderFee(ctx, args)
			if err != nil {
				log.Error("Error creating refund", zap.Error(err))
				continue
			}
		}
		time.Sleep(100 * time.Millisecond) // prevent  rate limit
	}

	return nil
}
