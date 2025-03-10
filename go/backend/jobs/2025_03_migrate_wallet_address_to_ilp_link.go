package jobs

import (
	"context"
	"errors"
	"os"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/uptrace/opentelemetry-go-extra/otelsql"
	"github.com/uptrace/opentelemetry-go-extra/otelsqlx"
	"gitlab.com/fynbos/log"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
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

	if err := executeActivity(ctx, a.UpdateBackendWalletRootToIlpActivity, "backend"); err != nil {
		return err
	}
	if err := executeActivity(ctx, a.UpdateRafikiWalletRootToIlpActivity, "rafiki"); err != nil {
		return err
	}
	if err := executeActivity(ctx, a.UpdateRafikiAuthWalletRootToIlpActivity, "rafiki auth"); err != nil {
		return err
	}
	logger.Info("Completed job MigrateWalletAddressesToIlpLinkJob update wallet to ilp.link")
	return nil
}

func executeActivity(ctx workflow.Context, activityFunc interface{}, activityName string) error {
	logger := workflow.GetLogger(ctx)
	err := workflow.ExecuteActivity(ctx, activityFunc).Get(ctx, nil)
	if err != nil {
		logger.Error("Migrate ["+activityName+"] to ilp.link failed", "Error", err)
		return err
	}
	logger.Info("Migrate [" + activityName + "] to ilp.link completed")
	return nil
}

func (a *Activity) UpdateBackendWalletRootToIlpActivity(ctx context.Context) error {
	log.Info("Starting [backend] wallet update to ilp.link")

	err := ExecuteTransaction(a.b.DB(), []string{
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
			"SET receiver_id = replace(receiver_id,'fynbos.me', 'ilp.link'); ",

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
	})
	if err != nil {
		log.Error("Error updating [backend] wallet to ilp.link: %v", zap.Error(err))
		return err
	}
	log.Info("Completed [backend] wallet update to ilp.link")
	return nil
}

func (a *Activity) UpdateRafikiWalletRootToIlpActivity(ctx context.Context) error {
	log.Info("Starting [rafiki] wallet update to ilp.link")
	connString := os.Getenv("RAFIKI_DB_URL")
	log.Info("Connection string: %v", zap.String("connString", connString))
	db, err := DbConnection(connString)
	if err != nil {
		log.Error("Error establishing db connection: %v", zap.Error(err))
		return err
	}

	trErr := ExecuteTransaction(db, []string{
		"UPDATE authServers " +
			"SET url = replace(url,'fynbos.me', 'ilp.link');",

		"UPDATE \"incomingPayments\" " +
			"SET client = replace(client,'fynbos.me', 'ilp.link');",

		"UPDATE quotes " +
			"SET client = replace(client,'fynbos.me', 'ilp.link'), " +
			"receiver = replace(receiver,'fynbos.me', 'ilp.link');",

		"UPDATE \"walletAddresses\" " +
			"SET url = replace(url,'fynbos.me', 'ilp.link');",

		"UPDATE webhookEvents " +
			"SET data = REPLACE(data::TEXT, 'fynbos.me', 'ilp.link')::JSON " +
			"WHERE data::TEXT LIKE '%fynbos.me%' ;",
	})

	db.Close()
	if trErr != nil {
		return trErr
	}
	return nil
}

func (a *Activity) UpdateRafikiAuthWalletRootToIlpActivity(ctx context.Context) error {
	log.Info("Starting [rafiki auth] wallet update to ilp.link")
	connString := os.Getenv("RAFIKI_AUTH_DB_URL")

	db, err := DbConnection(connString)
	if err != nil {
		log.Error("Error establishing db connection: %v", zap.Error(err))
		return err
	}

	trErr := ExecuteTransaction(db, []string{
		"UPDATE accesses " +
			"SET identifier = replace(identifier,'fynbos.me', 'ilp.link');",

		"UPDATE grants " +
			"SET client = replace(client,'fynbos.me', 'ilp.link');",
	})
	db.Close()
	if trErr != nil {
		return trErr
	}
	return nil
}

func DbConnection(connString string) (*sqlx.DB, error) {
	if connString == "" {
		log.Error("DB connection string is empty")
		return nil, errors.New("DB connection string is empty")
	}
	log.Info("Establishing db connection")
	db, err := otelsqlx.Connect("postgres", connString, otelsql.WithAttributes(semconv.DBSystemCockroachdb), otelsql.WithDBName("cockroachdb"))
	if err == nil {
		return db, nil
	} else {
		log.Error("Failed to connect to the database", zap.Error(err))
		defer db.Close()
		return nil, err
	}

}

func ExecuteTransaction(db *sqlx.DB, updates []string) error {
	if len(updates) == 0 {
		log.Error("No updates to execute")
		return nil
	}
	tx, err := db.Beginx()
	if err != nil {
		log.Error("Error starting wallet update transaction: %v", zap.Error(err))
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
	return nil
}
