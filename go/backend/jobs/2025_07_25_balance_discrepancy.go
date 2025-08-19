package jobs

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"math"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/google/uuid"
	"gitlab.com/fynbos/backend/currency"
	"gitlab.com/fynbos/backend/providers/gatehub"
	gatehub_external "gitlab.com/fynbos/backend/providers/gatehub/external"
	httplogger "gitlab.com/fynbos/backend/providers/http"
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
					continue
				}

				externalUserID, err := getExternalUserID(ctx, a.b, wallet.ID)
				if err != nil {
					continue
				}

				ebal, err := gatehubExternal.GetWalletBalances(ctx, externalUserID, la.ProviderID)
				if err != nil {
					continue
				}

				var balance, externalBalance float64
				if len(ebal) != 0 {
					balance, err = strconv.ParseFloat(bal.Available.FormatAmount(), 64)
					if err != nil {
						continue
					}
					externalBalance, err = strconv.ParseFloat(ebal[0].Available, 64)
					if err != nil {
						continue
					}

				}

				if externalBalance < balance {
					accs, err := a.b.Pacioli().GetAccounts(ctx, []string{la.ID})
					if err != nil || len(accs) == 0 {
						continue
					}

					uts, err := gatehubExternal.GetUserTransactions(ctx, externalUserID)
					if err != nil {
						continue
					}
					var feesTotal float64 = 0

					for _, ut := range uts {
						tfee, err := strconv.ParseFloat(ut.Fee, 64)
						if err != nil {
							continue
						}
						feesTotal += tfee
					}

					// backend transactions
					// hardcoded 1% fee
					_, err = a.b.DB().ExecContext(ctx, "UPDATE transactions SET provider_fee = amount / 100, amount = amount - (amount / 100), updated_at=now() WHERE provider_fee=0 AND type='deposit' AND provider='gatehub' AND wallet_id=$1;", wallet.ID)
					if err != nil {
						return err
					}

					//update pacioli side
					timeout := time.Hour * 24 * 365
					txID := uuid.NewString()
					feeAmount := currency.Amount{
						Value:    uint64(math.Round(feesTotal * 100)), // EUR scale = 2
						Currency: "EUR",
					}
					_, err = a.b.Gatehub().ReserveBalance(ctx, la.ID, txID, feeAmount, timeout)
					if err != nil {
						return err
					}
					err = a.b.Gatehub().FinaliseReserve(ctx, txID)
					if err != nil {
						return err
					}
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
