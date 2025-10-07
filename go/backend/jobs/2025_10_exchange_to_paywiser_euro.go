package jobs

import (
	"context"
	"net/http"
	"os"
	"time"

	"gitlab.com/fynbos/backend/providers/gatehub"

	"gitlab.com/fynbos/backend/providers/gatehub/external"
	httplogger "gitlab.com/fynbos/backend/providers/http"
	"gitlab.com/fynbos/log"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"
)

func BackfillPaywiserAccountsJob(ctx workflow.Context) (string, error) {
	var a *Activity
	ao := workflow.ActivityOptions{
		StartToCloseTimeout: 5 * time.Minute,
		RetryPolicy: &temporal.RetryPolicy{
			MaximumAttempts: 1,
		},
	}
	ctx = workflow.WithActivityOptions(ctx, ao)

	var gatehubWallets []string

	err := workflow.ExecuteActivity(ctx, a.GetGatehubUsersWalletIDs).Get(ctx, &gatehubWallets)
	if err != nil {
		return "", err
	}

	err = workflow.ExecuteActivity(ctx, a.BackfillPaywiserBalance, gatehubWallets).Get(ctx, nil)
	if err != nil {
		return "", err
	}

	return "BackfillPaywiserBalance done", nil
}

func (a *Activity) BackfillPaywiserBalance(ctx context.Context, gatehubWallets []string) error {
	ec := external.NewClient(
		os.Getenv("GATEHUB_APP_ID"),
		os.Getenv("GATEHUB_SECRET"),
		os.Getenv("GATEHUB_GATEWAY_ID"),
		&http.Client{
			Transport: otelhttp.NewTransport(
				httplogger.NewTransport(http.DefaultTransport, a.b, nil),
			),
		})
	for _, gw := range gatehubWallets {
		if gw == "" {
			continue
		}
		la, err := a.b.LinkedAccounts().ListByWalletId(ctx, gw)
		if err != nil {
			log.Error("linked account not found", zap.String("wallet_id", gw))
		}
		for _, l := range la {
			if l.Provider == gatehub.ProviderName && l.Type == gatehub.AccTypeBalance {

				balance, err := a.b.Gatehub().GetBalance(ctx, l.ID)
				if err != nil {
					log.Error("error getting balance for backfill transaction", zap.String("wallet_id", gw), zap.String("linked_account_id", l.ID))
				}
				transfer := balance.Total.Float64()
				if transfer <= 0 {
					log.Info("no balance to transfer", zap.String("wallet_id", gw), zap.String("linked_account_id", l.ID))
					continue
				}
				externalTx, err := ec.CreateTransaction(ctx, external.CreateTransactionRequest{
					SendingUserID:    "febc35fc-b48b-4db0-9066-2c3198aa9a0f",
					SendingAddress:   "520010820",
					ReceivingAddress: l.ProviderID,
					Amount:           transfer,
					Message:          "backfill transfer for sandbox",
					Type:             external.TransactionTypeHosted,
					VaultID:          ec.GetVaultID(),
				})

				if err != nil {
					log.Warn("error creating backfill transaction", zap.String("wallet_id", gw), zap.String("linked_account_id", l.ID), zap.Error(err))
				}
				log.Info("created external transaction", zap.Any("externalTx", externalTx), zap.Float64("amount", transfer))
				break
			}
		}
	}

	return nil
}

func (a *Activity) GetGatehubUsersWalletIDs(ctx context.Context) ([]string, error) {
	var gatehubWallets []string
	err := a.b.DB().SelectContext(ctx, &gatehubWallets, "SELECT wallet_id FROM gatehub_users")
	if err != nil {
		return nil, err
	}
	return gatehubWallets, nil
}
