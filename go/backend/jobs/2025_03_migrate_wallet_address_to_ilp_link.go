package jobs

import (
	"context"
	"errors"
	"os"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/uptrace/opentelemetry-go-extra/otelsql"
	"github.com/uptrace/opentelemetry-go-extra/otelsqlx"
	"gitlab.com/fynbos/env"
	"gitlab.com/fynbos/log"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"
)

type DomainInfo struct {
	OldDomain    string
	NewDomain    string
	OldAppDomain string
	NewAppDomain string
}

func MigrateWalletAddressesToIlpLinkJob(ctx workflow.Context) error {
	var domainInfo DomainInfo

	if env.IsLocal() {
		domainInfo = DomainInfo{
			OldDomain:    "fynbos.me",
			NewDomain:    "ilp.link",
			OldAppDomain: "fynbos.app",
			NewAppDomain: "interledger.app",
		}
	} else if env.IsDev() {
		domainInfo = DomainInfo{
			OldDomain:    "eu1.fynbos.me",
			NewDomain:    "sb.ilp.link",
			OldAppDomain: "eu1.fynbos.dev",
			NewAppDomain: "sb.interledger.app",
		}
	} else if env.IsProd() {
		domainInfo = DomainInfo{
			OldDomain:    "fynbos.me",
			NewDomain:    "ilp.link",
			OldAppDomain: "fynbos.app",
			NewAppDomain: "interledger.app",
		}
	}

	var a *Activity
	ao := workflow.ActivityOptions{
		StartToCloseTimeout: 10 * time.Second,
		RetryPolicy: &temporal.RetryPolicy{
			MaximumAttempts: 3,
		},
	}
	ctx = workflow.WithActivityOptions(ctx, ao)
	logger := workflow.GetLogger(ctx)
	logger.Info("Starting job MigrateWalletAddressesToIlpLinkJob update wallet to " + domainInfo.NewDomain)

	if err := executeActivity(ctx, a.UpdateBackendWalletRootToIlpActivity, "backend", domainInfo); err != nil {
		return err
	}
	if err := executeActivity(ctx, a.UpdateRafikiWalletRootToIlpActivity, "rafiki", domainInfo); err != nil {
		return err
	}
	if err := executeActivity(ctx, a.UpdateRafikiAuthWalletRootToIlpActivity, "rafiki auth", domainInfo); err != nil {
		return err
	}
	logger.Info("Completed job MigrateWalletAddressesToIlpLinkJob update wallet to " + domainInfo.NewDomain)
	return nil
}

func executeActivity(ctx workflow.Context, activityFunc interface{}, activityName string, domainInfo DomainInfo) error {
	logger := workflow.GetLogger(ctx)
	err := workflow.ExecuteActivity(ctx, activityFunc, domainInfo).Get(ctx, nil)
	if err != nil {
		logger.Error("Migrate ["+activityName+"] to "+domainInfo.NewDomain+" failed", "Error", err)
		return err
	}
	logger.Info("Migrate [" + activityName + "] to " + domainInfo.NewDomain + " completed")
	return nil
}

func (a *Activity) UpdateBackendWalletRootToIlpActivity(ctx context.Context, domainInfo DomainInfo) error {
	log.Info("Starting [backend] wallet update to " + domainInfo.NewDomain)

	queries := []struct {
		query string
		args  []interface{}
	}{
		{"UPDATE openpayments_incoming_payment SET sender_wallet_address = replace(sender_wallet_address, $1, $2) WHERE sender_wallet_address ILIKE '%' || $1 || '%';", []interface{}{domainInfo.OldDomain, domainInfo.NewDomain}},
		{"UPDATE openpayments_incoming_payment SET receiver_wallet_address = replace(receiver_wallet_address, $1, $2) WHERE receiver_wallet_address ILIKE '%' || $1 || '%';", []interface{}{domainInfo.OldDomain, domainInfo.NewDomain}},
		{"UPDATE openpayments_incoming_payment SET created_by = replace(created_by, $1, $2) WHERE created_by ILIKE '%' || $1 || '%';", []interface{}{domainInfo.OldDomain, domainInfo.NewDomain}},

		{"UPDATE openpayments_outgoing_payment SET sender_wallet_address = replace(sender_wallet_address, $1, $2) WHERE sender_wallet_address ILIKE '%' || $1 || '%';", []interface{}{domainInfo.OldDomain, domainInfo.NewDomain}},
		{"UPDATE openpayments_outgoing_payment SET receiver_wallet_address = replace(receiver_wallet_address, $1, $2) WHERE receiver_wallet_address ILIKE '%' || $1 || '%';", []interface{}{domainInfo.OldDomain, domainInfo.NewDomain}},
		{"UPDATE openpayments_outgoing_payment SET created_by = replace(created_by, $1, $2) WHERE created_by ILIKE '%' || $1 || '%';", []interface{}{domainInfo.OldDomain, domainInfo.NewDomain}},

		{"UPDATE openpayments_quotes SET sender_wallet_address = replace(sender_wallet_address, $1, $2) WHERE sender_wallet_address ILIKE '%' || $1 || '%';", []interface{}{domainInfo.OldDomain, domainInfo.NewDomain}},
		{"UPDATE openpayments_quotes SET receiver_wallet_address = replace(receiver_wallet_address, $1, $2) WHERE receiver_wallet_address ILIKE '%' || $1 || '%';", []interface{}{domainInfo.OldDomain, domainInfo.NewDomain}},
		{"UPDATE openpayments_quotes SET created_by = replace(created_by, $1, $2) WHERE created_by ILIKE '%' || $1 || '%';", []interface{}{domainInfo.OldDomain, domainInfo.NewDomain}},

		{"UPDATE payments SET receiver_id = replace(receiver_id, $1, $2) WHERE receiver_id ILIKE '%' || $1 || '%';", []interface{}{domainInfo.OldDomain, domainInfo.NewDomain}},
		{"UPDATE contacts SET payment_pointer = replace(payment_pointer, $1, $2) WHERE payment_pointer ILIKE '%' || $1 || '%';", []interface{}{domainInfo.OldDomain, domainInfo.NewDomain}},
		{"UPDATE discord_authorizations SET redirect_url = replace(redirect_url, $1, $2) WHERE redirect_url ILIKE '%' || $1 || '%';", []interface{}{domainInfo.OldDomain, domainInfo.NewDomain}},

		{"UPDATE transactions SET source = replace(source, $1, $2) WHERE source ILIKE '%' || $1 || '%';", []interface{}{domainInfo.OldDomain, domainInfo.NewDomain}},
		{"UPDATE transactions SET destination_identity = replace(destination_identity, $1, $2) WHERE destination_identity ILIKE '%' || $1 || '%';", []interface{}{domainInfo.OldDomain, domainInfo.NewDomain}},
		{"UPDATE transactions SET destination = replace(destination, $1, $2) WHERE destination ILIKE '%' || $1 || '%';", []interface{}{domainInfo.OldDomain, domainInfo.NewDomain}},
		{"UPDATE transactions SET title = replace(title, $1, $2) WHERE title ILIKE '%' || $1 || '%';", []interface{}{domainInfo.OldDomain, domainInfo.NewDomain}},

		{"UPDATE wallet_addresses SET url = replace(url, $1, $2) WHERE url ILIKE '%' || $1 || '%';", []interface{}{domainInfo.OldDomain, domainInfo.NewDomain}},
		{"UPDATE wallet_keys SET name = replace(name, 'Fynbos', 'Interledger');", nil},

		{"UPDATE discord_authorizations SET redirect_url = replace(redirect_url, $1, $2) WHERE redirect_url ILIKE '%' || $1 || '%';", []interface{}{domainInfo.OldAppDomain, domainInfo.NewAppDomain}},
		{"UPDATE slack_authorizations SET redirect_url = replace(redirect_url, $1, $2) WHERE redirect_url ILIKE '%' || $1 || '%';", []interface{}{domainInfo.OldAppDomain, domainInfo.NewAppDomain}},
		{"UPDATE twitter_authorizations SET redirect_url = replace(redirect_url, $1, $2) WHERE redirect_url ILIKE '%' || $1 || '%';", []interface{}{domainInfo.OldAppDomain, domainInfo.NewAppDomain}},
	}

	err := ExecuteTransaction(a.b.DB(), queries)
	if err != nil {
		log.Error("Error updating [backend] wallet to "+domainInfo.NewDomain+": %v", zap.Error(err))
		return err
	}
	log.Info("Completed [backend] wallet update to " + domainInfo.NewDomain)
	return nil
}

func (a *Activity) UpdateRafikiWalletRootToIlpActivity(ctx context.Context, domainInfo DomainInfo) error {
	log.Info("Starting [rafiki] wallet update to " + domainInfo.NewDomain)
	connString := os.Getenv("RAFIKI_DB_URL")
	log.Info("Connection string: %v", zap.String("connString", connString))

	db, err := DbConnection(connString)
	if err != nil {
		log.Error("Error establishing db connection: %v", zap.Error(err))
		return err
	}
	defer db.Close() // Ensure the database connection is closed after the transaction

	queries := []struct {
		query string
		args  []interface{}
	}{
		{"UPDATE \"incomingPayments\" SET client = replace(client, $1, $2) WHERE client ILIKE '%' || $1 || '%';", []interface{}{domainInfo.OldDomain, domainInfo.NewDomain}},
		{"UPDATE quotes SET receiver = replace(receiver, $1, $2) WHERE receiver ILIKE '%' || $1 || '%';", []interface{}{domainInfo.OldDomain, domainInfo.NewDomain}},
		{"UPDATE quotes SET client = replace(client, $1, $2) WHERE client ILIKE '%' || $1 || '%';", []interface{}{domainInfo.OldDomain, domainInfo.NewDomain}},
		{"UPDATE \"walletAddresses\" SET url = replace(url, $1, $2) WHERE url ILIKE '%' || $1 || '%';", []interface{}{domainInfo.OldDomain, domainInfo.NewDomain}},
		{"UPDATE \"webhookEvents\" SET data = REPLACE(data::TEXT, $1, $2)::JSON WHERE data::TEXT LIKE '%' || $1 || '%';", []interface{}{domainInfo.OldDomain, domainInfo.NewDomain}},
	}

	trErr := ExecuteTransaction(db, queries)
	if trErr != nil {
		return trErr
	}
	return nil
}

func (a *Activity) UpdateRafikiAuthWalletRootToIlpActivity(ctx context.Context, domainInfo DomainInfo) error {
	log.Info("Starting [rafiki auth] wallet update to " + domainInfo.NewDomain)
	connString := os.Getenv("RAFIKI_AUTH_DB_URL")

	db, err := DbConnection(connString)
	if err != nil {
		log.Error("Error establishing db connection: %v", zap.Error(err))
		return err
	}
	defer db.Close() // Ensure the database connection is closed after the transaction

	queries := []struct {
		query string
		args  []interface{}
	}{
		{"UPDATE accesses SET identifier = replace(identifier, $1, $2) WHERE identifier ILIKE '%' || $1 || '%';", []interface{}{domainInfo.OldDomain, domainInfo.NewDomain}},
		{"UPDATE grants SET client = replace(client, $1, $2) WHERE client ILIKE '%' || $1 || '%';", []interface{}{domainInfo.OldDomain, domainInfo.NewDomain}},
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
