package db

import (
	"context"
	"embed"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	_ "github.com/golang-migrate/migrate/v4/database/cockroachdb"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/google/uuid"
	"github.com/interledger/interledger-app/go/log"
	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"
)

const testingCrdbConnectionString = "postgres://postgres:password@127.0.0.1:55432/%s?sslmode=disable"

//go:embed schema.hcl
var schemaFile embed.FS

func Migrate(ctx context.Context, connString string) error {
	_, err := exec.LookPath("atlas")
	if err != nil {
		return err
	}

	schemaSQL, err := schemaFile.ReadFile("schema.hcl")
	if err != nil {
		return fmt.Errorf("failed to read embedded schema file: %w", err)
	}

	// Write the schemaSQL to a temporary file
	tmpFile, err := os.CreateTemp("", "schema-*.hcl")
	if err != nil {
		return fmt.Errorf("failed to create temporary file: %w", err)
	}
	defer tmpFile.Close()
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.Write(schemaSQL); err != nil {
		return fmt.Errorf("failed to write to temporary file: %w", err)
	}

	log.Info("Applying pacioli migrations")

	args := []string{
		"schema",
		"apply",
		"--auto-approve",
		"-u",
		connString,
		"--schema", "public",
		"-f",
		tmpFile.Name(),
	}

	out, err := exec.CommandContext(ctx, "atlas", args...).CombinedOutput()
	log.Info("atlas output", zap.String("out", fmt.Sprintf("out: %s", out)))
	if err != nil {
		log.Error("atlas error", zap.Error(err))
		return err
	}

	log.Info("Pacioli Atlas migration applied successfully", zap.String("output", string(out)))

	return nil
}

func MigrateTestDB(t *testing.T, ctx context.Context) (string, *sqlx.DB) {
	_, moduleDir, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("Could not get directory path for utils/testing.")
	}

	dbName := "pacioli_test_" + strings.Replace(uuid.NewString(), "-", "", 4)
	baseString := os.Getenv("DB_URL")
	if baseString == "" {
		baseString = testingCrdbConnectionString
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

	_, err = exec.LookPath("atlas")
	if err != nil {
		t.Fatal(err)
	}

	cmd := exec.CommandContext(ctx, "atlas", "schema", "apply", "--auto-approve", "-u", connString, "-f", filepath.Join(moduleDir, "../schema.hcl"))
	if err = cmd.Run(); err != nil {
		t.Fatal(err)
	}

	return connString, conn
}
