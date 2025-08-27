package jobs

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"

	"go.uber.org/zap"

	"github.com/google/uuid"
	"gitlab.com/fynbos/backend/currency"
	"gitlab.com/fynbos/backend/providers/gatehub"
	gatehub_external "gitlab.com/fynbos/backend/providers/gatehub/external"
	httplogger "gitlab.com/fynbos/backend/providers/http"
	"gitlab.com/fynbos/log"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

func BalanceDiscrepanciesJob(ctx workflow.Context) error {
	var a *Activity
	ao := workflow.ActivityOptions{
		StartToCloseTimeout: 10 * time.Second,
		RetryPolicy: &temporal.RetryPolicy{
			MaximumAttempts: 1,
		},
	}
	ctx = workflow.WithActivityOptions(ctx, ao)

	err := workflow.ExecuteActivity(ctx, a.BalanceDiscrepancies).Get(ctx, nil)
	if err != nil {
		return err
	}

	return nil
}

func (a *Activity) BalanceDiscrepancies(ctx context.Context) error {

	gatehubExternal := gatehub_external.NewClient(
		os.Getenv("GATEHUB_APP_ID"),
		os.Getenv("GATEHUB_SECRET"),
		&http.Client{
			Transport: otelhttp.NewTransport(
				httplogger.NewTransport(http.DefaultTransport, a.b, nil),
			),
		})

	type WalletsType struct {
		ID   string `db:"id"`
		NAME string `db:"name"`
	}

	var AllWallets []WalletsType
	err := a.b.DB().SelectContext(ctx, &AllWallets, "SELECT id, name FROM wallets")
	if err != nil {
		return err
	}

	for _, wallet := range AllWallets {

		lal, err := a.b.LinkedAccounts().ListByWalletId(ctx, wallet.ID)
		if err != nil {
			fmt.Println("Error la:", err)
			return err
		}

		for _, la := range lal {

			if la.Provider == gatehub.ProviderName && la.Type == gatehub.AccTypeBalance && la.DeletedAt.Time.IsZero() {
				bal, err := a.b.Gatehub().GetBalance(ctx, la.ID)
				if err != nil {
					return err
				}
				externalUserID, err := getExternalUserID(ctx, a.b, wallet.ID)
				if err != nil {
					continue
				}

				tx, err := a.b.DB().BeginTxx(ctx, nil)
				if err != nil {
					return err
				}

				defer func() {
					_ = tx.Rollback()
				}()

				// update deposits //1% fee
				_, err = tx.ExecContext(ctx, "UPDATE transactions SET provider_fee = ROUND(amount / 100,0), updated_at=now() WHERE provider_fee=0 AND type='deposit' AND provider='gatehub' AND wallet_id=$1;", wallet.ID)
				if err != nil {
					return err
				}

				// update withdrawals should add the 1// 1 euro fee
				_, err = tx.ExecContext(ctx, "UPDATE transactions SET provider_fee = 100, updated_at=now() WHERE provider_fee=0 AND type='withdrawal' AND provider='gatehub' AND wallet_id=$1;", wallet.ID)
				if err != nil {
					return err
				}

				//sum of fees and amount
				type Totals struct {
					Amount uint64 `db:"total_amount"`
					Fees   uint64 `db:"total_fees"`
				}
				var totals Totals
				err = tx.GetContext(ctx, &totals, `
					SELECT 
						COALESCE(SUM(
							CASE 
								WHEN type IN ('web_monetization_outgoing', 'withdrawal', 'open_payments_outgoing', 'sent') 
									THEN -amount
								ELSE amount
							END
						), 0) AS total_amount, 
						COALESCE(SUM(provider_fee), 0) AS total_fees 
					FROM transactions WHERE state='Completed' AND wallet_id=$1;`, wallet.ID)

				if err != nil {
					return err
				}

				// ignore users with 0 balance
				if totals.Amount <= 0 || totals.Fees <= 0 {
					log.Debug("No amount found for wallet:", zap.Uint64("fee", totals.Fees), zap.Uint64("amount", totals.Amount), zap.String("wallet", wallet.ID))
					continue
				}

				// get provider balance
				ebal, err := gatehubExternal.GetWalletBalances(ctx, externalUserID, la.ProviderID)
				if err != nil {
					continue
				}

				var externalBalance uint64
				if len(ebal) != 0 {
					externalBalanceFloat, err := strconv.ParseFloat(ebal[0].Available, 64)
					if err != nil {
						continue
					}
					externalBalance = uint64(externalBalanceFloat * 100)
				}

				// sanity check
				userBalance := uint64(totals.Amount - totals.Fees)
				if bal.Total.Value == userBalance {
					_ = tx.Rollback()
					continue
				}

				if externalBalance == userBalance {
					//happy we match provider records
					err = tx.Commit()
					if err != nil {
						return err
					}

					//update pacioli side
					timeout := time.Hour * 24 * 365
					txID := uuid.NewString()
					feeAmount := currency.Amount{
						Value:    uint64(totals.Fees), // EUR scale = 2
						Currency: "EUR",
					}
					log.Info("adding fees", zap.String("wallet_id", wallet.ID), zap.String("linked_account", la.ID), zap.Uint64("fee_amount", feeAmount.Value))
					_, err = a.b.Gatehub().ReserveBalance(ctx, la.ID, txID, feeAmount, timeout)
					if err != nil {
						return err
					}
					err = a.b.Gatehub().FinaliseReserve(ctx, txID)
					if err != nil {
						return err
					}
					log.Info("added fees", zap.String("wallet_id", wallet.ID), zap.String("linked_account", la.ID), zap.Uint64("fee_amount", feeAmount.Value))
				} else {
					log.Error("balances do not match", zap.String("wallet", wallet.ID), zap.Uint64("external_balance", externalBalance), zap.Uint64("current_balance", userBalance))
					_ = tx.Rollback()
				}

				time.Sleep(100 * time.Millisecond) // prevent  rate limit
			}

		}

	}

	return nil
}

func getExternalUserID(ctx context.Context, b Backends, walletID string) (string, error) {
	var externalID string
	err := b.DB().GetContext(ctx, &externalID, "SELECT external_id FROM gatehub_users WHERE wallet_id=$1;", walletID)
	if errors.Is(err, sql.ErrNoRows) {
		return "", fmt.Errorf("%w %s", gatehub.ErrNotFound, err)
	} else if err != nil {
		return "", fmt.Errorf("%w %s", gatehub.ErrInternal, err)
	}

	return externalID, err
}
