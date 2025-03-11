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

const oldDomain = "fynbos.me"
const newDomain = "ilp.link"
const oldAppDomain = "fynbos.app"
const newAppDomain = "interledger.app"

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
	logger.Info("Starting job MigrateWalletAddressesToIlpLinkJob migrate wallet to " + newDomain)

	if err := executeActivity(ctx, a.UpdateBackendWalletRootToIlpActivity, "backend"); err != nil {
		return err
	}
	if err := executeActivity(ctx, a.UpdateRafikiWalletRootToIlpActivity, "rafiki"); err != nil {
		return err
	}
	if err := executeActivity(ctx, a.UpdateRafikiAuthWalletRootToIlpActivity, "rafiki auth"); err != nil {
		return err
	}
	logger.Info("Completed job MigrateWalletAddressesToIlpLinkJob migrate wallet to " + newDomain)
	return nil
}

func executeActivity(ctx workflow.Context, activityFunc interface{}, activityName string) error {
	logger := workflow.GetLogger(ctx)
	err := workflow.ExecuteActivity(ctx, activityFunc).Get(ctx, nil)
	if err != nil {
		logger.Error("Migrate ["+activityName+"] to "+newDomain+" failed", "Error", err)
		return err
	}
	logger.Info("Migrate [" + activityName + "] to " + newDomain + " completed")
	return nil
}

func (a *Activity) UpdateBackendWalletRootToIlpActivity(ctx context.Context) error {
	log.Info("Starting [backend] wallet update to " + newDomain)

	queries := []struct {
		query string
		args  []interface{}
	}{
		{"UPDATE openpayments_incoming_payment SET sender_wallet_address = replace(sender_wallet_address, $1, $2) WHERE sender_wallet_address ILIKE '%' || $1 || '%';", []interface{}{oldDomain, newDomain}},
		{"UPDATE openpayments_incoming_payment SET receiver_wallet_address = replace(receiver_wallet_address, $1, $2) WHERE receiver_wallet_address ILIKE '%' || $1 || '%';", []interface{}{oldDomain, newDomain}},
		{"UPDATE openpayments_incoming_payment SET created_by = replace(created_by, $1, $2) WHERE created_by ILIKE '%' || $1 || '%';", []interface{}{oldDomain, newDomain}},

		{"UPDATE openpayments_outgoing_payment SET sender_wallet_address = replace(sender_wallet_address, $1, $2) WHERE sender_wallet_address ILIKE '%' || $1 || '%';", []interface{}{oldDomain, newDomain}},
		{"UPDATE openpayments_outgoing_payment SET receiver_wallet_address = replace(receiver_wallet_address, $1, $2) WHERE receiver_wallet_address ILIKE '%' || $1 || '%';", []interface{}{oldDomain, newDomain}},
		{"UPDATE openpayments_outgoing_payment SET created_by = replace(created_by, $1, $2) WHERE created_by ILIKE '%' || $1 || '%';", []interface{}{oldDomain, newDomain}},

		{"UPDATE openpayments_quotes SET sender_wallet_address = replace(sender_wallet_address, $1, $2) WHERE sender_wallet_address ILIKE '%' || $1 || '%';", []interface{}{oldDomain, newDomain}},
		{"UPDATE openpayments_quotes SET receiver_wallet_address = replace(receiver_wallet_address, $1, $2) WHERE receiver_wallet_address ILIKE '%' || $1 || '%';", []interface{}{oldDomain, newDomain}},
		{"UPDATE openpayments_quotes SET created_by = replace(created_by, $1, $2) WHERE created_by ILIKE '%' || $1 || '%';", []interface{}{oldDomain, newDomain}},

		{"UPDATE payments SET receiver_id = replace(receiver_id, $1, $2) WHERE receiver_id ILIKE '%' || $1 || '%';", []interface{}{oldDomain, newDomain}},
		{"UPDATE contacts SET payment_pointer = replace(payment_pointer, $1, $2) WHERE payment_pointer ILIKE '%' || $1 || '%';", []interface{}{oldDomain, newDomain}},
		{"UPDATE discord_authorizations SET redirect_url = replace(redirect_url, $1, $2) WHERE redirect_url ILIKE '%' || $1 || '%';", []interface{}{oldDomain, newDomain}},

		{"UPDATE transactions SET source = replace(source, $1, $2) WHERE source ILIKE '%' || $1 || '%';", []interface{}{oldDomain, newDomain}},
		{"UPDATE transactions SET destination_identity = replace(destination_identity, $1, $2) WHERE destination_identity ILIKE '%' || $1 || '%';", []interface{}{oldDomain, newDomain}},
		{"UPDATE transactions SET destination = replace(destination, $1, $2) WHERE destination ILIKE '%' || $1 || '%';", []interface{}{oldDomain, newDomain}},

		{"UPDATE wallet_addresses SET url = replace(url, $1, $2) WHERE url ILIKE '%' || $1 || '%';", []interface{}{oldDomain, newDomain}},
		{"UPDATE wallet_keys SET name = replace(name, 'Fynbos', 'Interledger') WHERE name ILIKE '%Fynbos%';", nil},

		{"UPDATE public.slack_authorizations SET redirect_url = replace(redirect_url, $1, $2) WHERE redirect_url ILIKE '%' || $1 || '%';", []interface{}{oldAppDomain, newAppDomain}},
		{"UPDATE public.twitter_authorizations SET redirect_url = replace(redirect_url, $1, $2) WHERE redirect_url ILIKE '%' || $1 || '%';", []interface{}{oldAppDomain, newAppDomain}},
	}

	err := ExecuteTransaction(a.b.DB(), queries)
	if err != nil {
		log.Error("Error updating [backend] wallet to "+newDomain+": %v", zap.Error(err))
		return err
	}
	log.Info("Completed [backend] wallet update to " + newDomain)
	return nil
}

func (a *Activity) UpdateRafikiWalletRootToIlpActivity(ctx context.Context) error {
	log.Info("Starting [rafiki] wallet update to " + newDomain)
	connString := os.Getenv("RAFIKI_DB_URL")
	log.Info("Connection string: %v", zap.String("connString", connString))

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
		{"UPDATE \"incomingPayments\" SET client = replace(client, $1, $2) WHERE client ILIKE '%' || $1 || '%';", []interface{}{oldDomain, newDomain}},
		{"UPDATE quotes SET client = replace(client, $1, $2), receiver = replace(receiver, $1, $2) WHERE receiver ILIKE '%' || $1 || '%';", []interface{}{oldDomain, newDomain}},
		{"UPDATE \"walletAddresses\" SET url = replace(url, $1, $2) WHERE url  ILIKE '%' || $1 || '%';", []interface{}{oldDomain, newDomain}},
		{"UPDATE \"webhookEvents\" SET data = REPLACE(data::TEXT, $1, $2)::JSON WHERE data::TEXT LIKE '%' || $1 || '%';", []interface{}{oldDomain, newDomain}},
	}

	trErr := ExecuteTransaction(db, queries)
	if trErr != nil {
		return trErr
	}
	return nil
}

func (a *Activity) UpdateRafikiAuthWalletRootToIlpActivity(ctx context.Context) error {
	log.Info("Starting [rafiki auth] wallet update to " + newDomain)
	connString := os.Getenv("RAFIKI_AUTH_DB_URL")

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
		{"UPDATE accesses SET identifier = replace(identifier, $1, $2) WHERE identifier  ILIKE '%' || $1 || '%';", []interface{}{oldDomain, newDomain}},
		{"UPDATE grants SET client = replace(client, $1, $2) WHERE client  ILIKE '%' || $1 || '%';", []interface{}{oldDomain, newDomain}},
	}

	trErr := ExecuteTransaction(db, queries)
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
	if err != nil {
		log.Error("Failed to connect to the database", zap.Error(err))
		return nil, err
	}
	return db, nil
}

func ExecuteTransaction(db *sqlx.DB, queries []struct {
	query string
	args  []interface{}
}) error {
	tx, err := db.BeginTxx(context.Background(), nil)
	if err != nil {
		log.Error("Error starting transaction", zap.Error(err))
		return err
	}

	for _, q := range queries {
		if q.args == nil {
			if _, err := tx.Exec(q.query); err != nil {
				tx.Rollback()
				log.Error("Error executing query", zap.Error(err))
				return err
			}
		} else {
			if _, err := tx.Exec(q.query, q.args...); err != nil {
				tx.Rollback()
				log.Error("Error executing query", zap.Error(err))
				return err
			}
		}
	}

	if err := tx.Commit(); err != nil {
		log.Error("Error committing transaction", zap.Error(err))
		return err
	}

	return nil
}
