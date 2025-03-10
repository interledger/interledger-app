package jobs

import (
	"context"
	"time"

	"gitlab.com/fynbos/log"
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"
)

func MigrateWalletAddressesToIlpLinkJob(ctx workflow.Context) error {
	var a *Activity
	ao := workflow.ActivityOptions{
		StartToCloseTimeout: 10 * time.Second,
		RetryPolicy: &temporal.RetryPolicy{
			MaximumAttempts: 3,
		},
	}
	ctx = workflow.WithActivityOptions(ctx, ao)
	logger := workflow.GetLogger(ctx)
	logger.Info("Starting job MigrateWalletAddressesToIlpLinkJob update wallet to ilp.link")
	err := workflow.ExecuteActivity(ctx, a.UpdateWalletRootToIlpActivity).Get(ctx, nil)
	if err != nil {
		logger.Error("Job update wallet to ilp.link failed", "Error", err)
		return err
	} else {
		logger.Info("Job update wallet to ilp.link completed")
	}

	return err
}

func (a *Activity) UpdateWalletRootToIlpActivity(ctx context.Context) error {
	log.Info("Starting wallet update to ilp.link")
	tx, err := a.b.DB().Beginx()
	if err != nil {
		log.Error("Error starting wallet update transaction: %v", zap.Error(err))
	}
	updates := []string{
		"UPDATE openpayments_incoming_payment " +
			"SET sender_wallet_address = replace(sender_wallet_address,'fynbos.me', 'ilp.link'), " +
			"receiver_wallet_address = replace(receiver_wallet_address,'fynbos.me', 'ilp.link'), " +
			"created_by = replace(created_by,'fynbos.me', 'ilp.link');",

		"UPDATE openpayments_outgoing_payment " +
			"SET sender_wallet_address = replace(sender_wallet_address,'fynbos.me', 'ilp.link'), " +
			"receiver_wallet_address = replace(receiver_wallet_address,'fynbos.me', 'ilp.link'), " +
			"created_by = replace(created_by,'fynbos.me', 'ilp.link');",

		"UPDATE openpayments_quotes " +
			"SET sender_wallet_address = replace(sender_wallet_address,'fynbos.me', 'ilp.link'), " +
			"receiver_wallet_address = replace(receiver_wallet_address,'fynbos.me', 'ilp.link'), " +
			"created_by = replace(created_by,'fynbos.me', 'ilp.link');",

		"UPDATE payments " +
			"SET receiver_id = replace(receiver_id,'fynbos.me', 'ilp.link'); " +
			"-- public_id = replace(public_id,'fynbos_', 'interledger_') ASK",

		"UPDATE contacts " +
			"SET payment_pointer = replace(payment_pointer,'fynbos.me', 'ilp.link');",

		"UPDATE discord_authorizations " +
			"SET redirect_url = replace(redirect_url,'fynbos.me', 'ilp.link');",

		"UPDATE transactions " +
			"SET source = replace(source,'fynbos.me', 'ilp.link'), " +
			"destination_identity = replace(destination_identity,'fynbos.me', 'ilp.link');",

		"UPDATE wallet_addresses " +
			"SET url = replace(url,'fynbos.me', 'ilp.link');",

		"UPDATE wallet_keys  " +
			"SET name =  replace(name,'Fynbos', 'Interledger');",
		// redirect form slack and twitter (functionalities not used at this point)
		"UPDATE public.slack_authorizations  " +
			"SET redirect_url  =  replace(redirect_url, 'fynbos.app', 'interledger.app');",

		"UPDATE  public.twitter_authorizations " +
			"SET redirect_url =  replace(redirect_url,'fynbos.app', 'interledger.app');",
	}

	for _, update := range updates {
		_, err := tx.Exec(update)
		if err != nil {
			tx.Rollback() // Rollback if any update fails
			log.Error("Error executing wallet update: %v", zap.Error(err))
			return err
		}
	}

	if err := tx.Commit(); err != nil {
		log.Error("Error committing wallet update transaction: %v", zap.Error(err))
	}

	return err
}
