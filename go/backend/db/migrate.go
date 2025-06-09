package db

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"gitlab.com/fynbos/env"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/uptrace/opentelemetry-go-extra/otelsql"
	"github.com/uptrace/opentelemetry-go-extra/otelsqlx"
	"gitlab.com/fynbos/log"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
	"go.uber.org/zap"
)

const testingConnectionString = "postgres://postgres:password@0.0.0.0:5432/%s?sslmode=disable"

func Migrate(ctx context.Context, connString string) error {
	_, moduleDir, _, ok := runtime.Caller(0)
	if !ok {
		return fmt.Errorf("Could not get directory path for utils/testing.")
	}

	f, err := os.Open(filepath.Join(moduleDir, "../schema.hcl"))
	if err != nil {
		log.Warn("Failed to store schema hash in database.", zap.Error(err))
	}
	defer f.Close()

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		log.Warn("Failed to store schema hash in database.", zap.Error(err))
	}
	schemaHash := base64.StdEncoding.EncodeToString(h.Sum(nil))

	var currentHash string
	db, err := otelsqlx.Connect("postgres", connString, otelsql.WithAttributes(semconv.DBSystemCockroachdb), otelsql.WithDBName("cockroachdb"))
	if err != nil {
		log.Warn("Failed to connect to database to determine currently deployed schema.", zap.Error(err))
	} else {
		defer db.Close()
		err = db.GetContext(ctx, &currentHash, "SELECT hash FROM atlas_schema_history ORDER BY created_at desc LIMIT 1;")
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			log.Warn("Failed to close db connection after storing schema hash.", zap.Error(err))
		}
	}

	if strings.EqualFold(currentHash, schemaHash) {
		log.Info("Schema already deployed.", zap.String("hash", currentHash))
		return seedSysAccounts(ctx, db)
	}

	_, err = exec.LookPath("atlas")
	if err != nil {
		return err
	}
	args := []string{
		"schema",
		"apply",
		"--auto-approve",
		"-u",
		connString,
		"--schema", "public",
		"-f",
		filepath.Join(moduleDir, "../schema.hcl"),
	}

	out, err := exec.CommandContext(ctx, "atlas", args...).CombinedOutput()
	if err != nil {
		log.Info("atlas output", zap.String("out", fmt.Sprintf("out: %s", out)))
		log.Error("error migrating", zap.Error(err))
		return err
	}

	log.Info("atlas output", zap.String("out", fmt.Sprintf("out: %s", out)))

	if db != nil {
		result, err := db.ExecContext(ctx, "INSERT INTO atlas_schema_history (hash) VALUES ($1);", schemaHash)
		if err != nil {
			log.Warn("Failed to update deployed schema hash.", zap.Error(err))
		}

		if rows, _ := result.RowsAffected(); rows < 1 {
			log.Warn("Failed to update deployed schema hash. No row inserted.", zap.String("hash", schemaHash))
		}
	}

	err = seedSysAccounts(ctx, db)
	
	if err == nil {
		log.Info("MIGRATION APPLIED for backend")
	}

	return err
}

func CreateExpIndex(ctx context.Context, db *sqlx.DB) error {
	_, err := db.ExecContext(ctx, waExpIndex)
	if err != nil {
		return err
	}

	return nil
}

func MigrateTestDB(t *testing.T, ctx context.Context) *sqlx.DB {
	_, moduleDir, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("Could not get directory path for utils/testing.")
	}

	dbName := "backend_test_" + strings.Replace(uuid.NewString(), "-", "", 4)
	baseString := os.Getenv("DB_URL")
	if baseString == "" {
		baseString = testingConnectionString
	}
	connString := fmt.Sprintf(baseString, "")
	db, err := sqlx.Connect("postgres", connString)
	if err != nil {
		t.Fatal(err)
	}

	query := fmt.Sprintf("CREATE DATABASE %s;", dbName)
	_, err = db.ExecContext(ctx, query)
	if err != nil {
		t.Fatal(err)
	}

	connString = fmt.Sprintf(baseString, dbName)
	conn, err := sqlx.Connect("postgres", connString)
	if err != nil {
		t.Fatal(err)
	}

	query = "CREATE EXTENSION IF NOT EXISTS pg_trgm;"
	_, err = conn.ExecContext(ctx, query)
	if err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() {
		if err = conn.Close(); err != nil {
			t.Fatal(err)
		}

		cleanupQuery := fmt.Sprintf("DROP DATABASE %s;", dbName)
		_, err := db.ExecContext(ctx, cleanupQuery)
		if err != nil {
			t.Fatal(err)
		}

		if err = db.Close(); err != nil {
			t.Fatal(err)
		}
	})

	var migrations string
	err = filepath.Walk(filepath.Join(moduleDir, "../testmigrations"), func(path string, info fs.FileInfo, err error) error {
		if strings.Contains(info.Name(), ".sql") {
			sql, err := os.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}

			migrations = migrations + "\n" + string(sql)
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if migrations == "" {
		t.Fatal("No migrations found.")
	}

	_, err = conn.ExecContext(ctx, migrations)
	if err != nil {
		t.Fatal(err)
	}

	_, err = conn.ExecContext(ctx, waExpIndex)
	if err != nil {
		t.Fatal(err)
	}

	err = seedSysAccounts(ctx, conn)
	if err != nil {
		t.Fatal(err)
	}

	return conn
}

func seedSysAccounts(ctx context.Context, dbc *sqlx.DB) error {
	log.Info("Seeding system accounts...")
	// Can't use linkedaccount/rafiki constants for cyclic dependencies
	_, err := dbc.ExecContext(ctx, fmt.Sprintf(seedSQL, env.OpenPaymentsURL()))

	return err
}

const waExpIndex = `
CREATE UNIQUE INDEX IF NOT EXISTS "wallet_address_url_lower" ON "public"."wallet_addresses" (lower(url));
`

const seedSQL = `
INSERT INTO wallets (id, name) VALUES ('31db2044-0b83-4aae-9cd8-b5b63cc85414', 'Web Monetization') ON CONFLICT DO NOTHING;
INSERT INTO wallet_addresses(id, url, wallet_id) VALUES ('c4916e50-9b8c-4b54-87fd-419b2e4daee8','%s/webmonetization', '31db2044-0b83-4aae-9cd8-b5b63cc85414') ON CONFLICT DO NOTHING;
INSERT INTO linked_accounts(id, wallet_id, name, mask, provider, provider_id, type, nickname, state, can_send, can_receive, send_currency, receive_currency) VALUES ('8d97fe51-543c-4577-9ab1-18078caa371d','31db2044-0b83-4aae-9cd8-b5b63cc85414', 'web_monetization', '1234', 'rafiki', '0000', 'rafiki_web_monetization', 'Web Monetization', 'Verified', true, false, 'USD','USD') ON CONFLICT (id)
do update set can_receive = false, send_currency = 'USD', receive_currency = 'USD';
INSERT INTO linked_accounts(id, wallet_id, name, mask, provider, provider_id, type, nickname, state, can_send, can_receive, send_currency, receive_currency) VALUES ('905e2c9b-a8b7-4cf2-9449-197e7029052e','31db2044-0b83-4aae-9cd8-b5b63cc85414', 'web_monetization_zar', '1234', 'xago', '0000', 'balance', 'Web Monetization ZAR', 'Verified', true, false, 'ZAR','ZAR') ON CONFLICT DO NOTHING;
`
