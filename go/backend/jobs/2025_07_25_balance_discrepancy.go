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

	"gitlab.com/fynbos/backend/providers/gatehub"
	gatehub_external "gitlab.com/fynbos/backend/providers/gatehub/external"
	httplogger "gitlab.com/fynbos/backend/providers/http"
	"gitlab.com/fynbos/log"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"
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

	// ebal, err := gatehubExternal.GetWalletBalances(ctx, "0326e583-1feb-4ff5-ad69-6e7368108cc2", "631861543")
	// ebal, err := gatehubExternal.GetTransaction(ctx, "0326e583-1feb-4ff5-ad69-6e7368108cc2", "dd86fc1a-3484-427b-8eee-e1877d7fe83a")
	// ebal, err := gatehubExternal.GetUserTransactions(ctx, "0326e583-1feb-4ff5-ad69-6e7368108cc2")

	// ebal, err := gatehubExternal.GetFeeValue(ctx, "5", "EUR")
	// if err != nil {
	// 	log.Error("Error in gatehub: %v", zap.Error(err))
	// 	return err
	// }
	// log.Info(fmt.Sprintf("HERE %s", ebal))

	// return nil

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
				// REMOVE THIS
				if externalUserID != "0326e583-1feb-4ff5-ad69-6e7368108cc2" || la.ProviderID != "631861543" {
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

					log.Info(fmt.Sprintf("Available %.2f", balance))
					log.Info(fmt.Sprintf("External %.2f", externalBalance))
				}

				if externalBalance < balance {
					accs, err := a.b.Pacioli().GetAccounts(ctx, []string{la.ID})
					if err != nil || len(accs) == 0 {
						continue
					}
					CreditAccountID := la.ID

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
					_, err = a.b.DB().ExecContext(ctx, "UPDATE transactions SET provider_fee=amount*0.01, amount=amount*0.99, updated_at=now() WHERE provider_fee=0 AND type='deposit' AND provider='gatehub' AND wallet_id=$1;", wallet.ID)
					if err != nil {
						return err
					}
					//update pacioli side
					args := &ActivityArgs{feeTotal: feesTotal, ID: CreditAccountID}
					if err := getFeesAndAdjustBalance(ctx, args); err != nil {
						return err
					}
				}

				time.Sleep(100 * time.Millisecond) // prevent  rate limit
			}

		}

	}

	return nil
}

type ActivityArgs struct {
	feeTotal float64
	ID       string
}

func getFeesAndAdjustBalance(ctx context.Context, fargs *ActivityArgs) error {
	log.Info("Starting pacioli update ")
	connString := os.Getenv("PACIOLI_DB_URL")

	db, err := DbConnection(connString)
	if err != nil {
		log.Error("Error establishing db connection: %v", zap.Error(err))
		return err
	}
	defer db.Close()

	queries := []struct {
		query string
		args  []interface{}
	}{
		{"UPDATE ledger_accounts SET credits_posted=credits_posted-$1, updated_at=now() WHERE id=$2;", []interface{}{fargs.feeTotal, fargs.ID}},
		{"UPDATE ledger_transfers SET provider_fee=amount*0.01, amount=amount*0.99, updated_at=now() WHERE provider_fee=0 AND credit_account_id=$1;", []interface{}{fargs.ID}},
	}

	trErr := ExecuteTransaction(db, queries)
	if trErr != nil {
		return trErr
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
