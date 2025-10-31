package jobs

import (
	"context"
	"errors"
	"time"

	"go.temporal.io/sdk/activity"
	"go.temporal.io/sdk/workflow"
)

func PtiDropOldDataWorkflow(ctx workflow.Context) (err error) {
	var a *Activity
	ao := workflow.ActivityOptions{
		StartToCloseTimeout: 5 * time.Minute,
	}
	ctx = workflow.WithActivityOptions(ctx, ao)

	logger := workflow.GetLogger(ctx)
	logger.Info("dropping PTI old data")

	if err = workflow.ExecuteActivity(ctx, a.PtiDropWalletOldDataActivity).Get(ctx, nil); err != nil {
		logger.Error("RegisterWalletsPrivateKeyActivity Activity failed.", "Error", err)
		return
	}

	logger.Info("PTI old data dropped")

	return
}

// this method is designed to work atomic
// all the apparent "redundat" execution is intended to achive ALL OR NOTHING consistency
func (a *Activity) PtiDropWalletOldDataActivity(ctx context.Context) (err error) {
	logger := activity.GetLogger(ctx)

	type WalletLinkedAccount struct {
		WalletId       string `db:"wallet_id"`
		LinkedAccoutId string `db:"linked_account_id"`
	}

	var wlas []WalletLinkedAccount

	q := `select 
			ws.id as wallet_id, 
			la.id as linked_account_id 
		from pti_users pti 
		join wallets ws on
			pti.wallet_id = ws.id
		join linked_accounts la on 
			ws.id = la.wallet_id and
			la.provider = 'pti';`
	if err = a.b.DB().SelectContext(ctx, &wlas, q); err != nil {
		logger.Error("failed to fetch data", err)
		return
	}

	if len(wlas) == 0 {
		// this might indicate something wrong, just bail asap
		// sanity check
		logger.Error("wallets/linkedaccounts empty")
		return errors.New("wallets/linkedaccounts empty")
	}

	tx, err := a.b.DB().Beginx()
	if err != nil {
		logger.Error("failed to create sql transaction")
		return
	}

	for _, wla := range wlas {
		// payments table
		if _, err := tx.Exec("delete from payments where sender_id = $1 and sender_account = $2 and sender_currency = 'USD';", wla.WalletId, wla.LinkedAccoutId); err != nil {
			goto rollback
		}

		// transactions table, in short
		// this query forces linked_account_id into transactions as is not present there
		qTransactions := `delete 
							tons 
						  from transactions tons
						  join transfers ters on
							tons.id = ters.transaction_id and 
							tons.asset_code = 'USD' and 
							ters.asset_code = 'USD'
						  where 
						  	tons.provider = 'pti' and 
						  	ters.linked_acc_id = $1;`
		if _, err := tx.Exec(qTransactions, wla.LinkedAccoutId); err != nil {
			logger.Error("transactions table", err)
			goto rollback
		}

		// transfers table
		if _, err := tx.Exec("delete from transfers where linked_acc_id = $1 and asset_code = 'USD';", wla.LinkedAccoutId); err != nil {
			logger.Error("transfers table", err)
			goto rollback
		}

		// pti_transactions table
		if _, err := tx.Exec("truncate pti_transactions;"); err != nil {
			logger.Error("pti_transactions table", err)
			goto rollback
		}

		// wallet_kyc_status
		if _, err := tx.Exec("update walet_kyc_status set status = 0 where wallet_id = $1;", wla.WalletId); err != nil {
			logger.Error("wallet_kyc_status table", err)
			goto rollback
		}

		// linked_accounts
		qLinkedAccounts := `delete 
							from linked_accounts where
								id = $1 and
								wallet_id = $2 and
								provider = 'pti' and
								send_country = 'US' and
								receive_country = 'US' and 
								send_currency = 'USD' and 
								receive_currency = 'USD';`
		if _, err := tx.Exec(qLinkedAccounts, wla.LinkedAccoutId, wla.WalletId); err != nil {
			goto rollback
		}

	}

rollback:
	err = tx.Rollback()
	if err != nil {
		logger.Error("failed to rollback", err) // ouch! the worst thing can happen
		return err
	}

	return tx.Commit() // fingers crossed
}
