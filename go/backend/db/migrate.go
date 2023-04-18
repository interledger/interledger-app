package db

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"gitlab.com/fynbos/log"
	"go.uber.org/zap"
)

const testingCrdbConnectionString = "postgres://root@0.0.0.0:26257/%s?sslmode=disable"

func Migrate(ctx context.Context, connString string) error {
	_, moduleDir, _, ok := runtime.Caller(0)
	if !ok {
		return fmt.Errorf("Could not get directory path for utils/testing.")
	}

	_, err := exec.LookPath("atlas")
	if err != nil {
		return err
	}
	args := []string{
		"schema",
		"apply",
		"--auto-approve",
		"--dev-url",
		connString,
		"-u",
		connString,
		"--exclude",
		"public.payment_pointers.payment_pointers_url_lower",
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

	return nil
}

func CreateExpIndex(ctx context.Context, db *sqlx.DB) error {
	_, err := db.ExecContext(ctx, ppExpIndex)
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
	connString := os.Getenv("DB_URL")
	if connString == "" {
		connString = testingCrdbConnectionString
	}
	connString = fmt.Sprintf(connString, dbName)
	db, err := sqlx.Connect("postgres", connString)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		cleanupQuery := fmt.Sprintf("DROP DATABASE %s;", dbName)
		_, err := db.ExecContext(ctx, cleanupQuery)
		if err != nil {
			t.Fatal(err)
		}

		if err = db.Close(); err != nil {
			t.Fatal(err)
		}
	})

	query := fmt.Sprintf("CREATE DATABASE %s;", dbName)
	_, err = db.ExecContext(ctx, query)
	if err != nil {
		t.Fatal(err)
	}

	_, err = exec.LookPath("atlas")
	if err != nil {
		t.Fatal(err)
	}

	out, err := exec.CommandContext(ctx,
		"atlas", "schema", "apply", "--auto-approve", "-u", connString,
		"--exclude", "public.payment_pointers.payment_pointers_url_lower",
		"-f", filepath.Join(moduleDir, "../schema.hcl")).CombinedOutput()
	if err != nil {
		log.Error("error migrating", zap.String("output", string(out)))
		t.Fatal(err)
	}

	_, err = db.ExecContext(ctx, ppExpIndex)
	if err != nil {
		t.Fatal(err)
	}

	return db
}

const ppExpIndex = `
CREATE UNIQUE INDEX IF NOT EXISTS "payment_pointers_url_lower" ON "public"."payment_pointers" (lower(url));
`
