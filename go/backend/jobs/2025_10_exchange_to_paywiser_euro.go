package jobs

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"

	"gitlab.com/fynbos/backend/linkedaccounts"
	"gitlab.com/fynbos/backend/providers/gatehub"

	"gitlab.com/fynbos/backend/providers/gatehub/external"
	httplogger "gitlab.com/fynbos/backend/providers/http"
	"gitlab.com/fynbos/log"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"
)

func BackfillPaywiserAccountsJob(ctx workflow.Context, walletID string) (string, error) {
	var a *Activity
	ao := workflow.ActivityOptions{
		RetryPolicy: &temporal.RetryPolicy{
			MaximumAttempts: 1,
		},
	}
	ctx = workflow.WithActivityOptions(ctx, ao)

	var gatehubWallets []string
	if walletID != "" {
		err := workflow.ExecuteActivity(ctx, a.GetGatehubUsersWalletIDs).Get(ctx, &gatehubWallets)
		if err != nil {
			return "", err
		}
	} else {
		gatehubWallets = append(gatehubWallets, walletID)
	}

	var la linkedaccounts.LinkedAccount
	err := workflow.ExecuteActivity(ctx, a.BackfillPaywiserBalance, gatehubWallets).Get(ctx, &la)
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
		la, err := a.b.LinkedAccounts().ListByWalletId(ctx, gw)
		if err != nil {
			log.Error("linked account not found", zap.String("wallet_id", gw))
		}
		for _, l := range la {
			if l.Provider == gatehub.ProviderName && l.Type == gatehub.AccTypeBalance {

				balance, err := a.b.Gatehub().GetBalance(ctx, l.ID)
				if err != nil {
					log.Error("error getting balance from gatehub", zap.String("wallet_id", gw), zap.String("linked_account_id", l.ID))
				}
				transfer := balance.Total.Float64()

				externalTx, err := ec.CreateTransaction(ctx, external.CreateTransactionRequest{
					SendingUserID:    "febc35fc-b48b-4db0-9066-2c3198aa9a0f",
					SendingAddress:   "520010820",
					ReceivingAddress: l.ProviderID,
					Amount:           transfer,
					Message:          "exchange for sandbox",
					Type:             external.TransactionTypeHosted,
					VaultID:          ec.GetVaultID(),
				})
				if errors.Is(err, external.ErrNotFound) {
					return fmt.Errorf("%w %s", gatehub.ErrNotFound, err)
				}
				if err != nil {
					return fmt.Errorf("%w %s", gatehub.ErrInternal, err)
				}
				log.Info("created external transaction", zap.String("transaction_id", externalTx.ID), zap.Float64("amount", transfer))
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
