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
	"github.com/jmoiron/sqlx"
	"gitlab.com/fynbos/log"
	"go.uber.org/zap"
)

const testingCrdbConnectionString = "postgres://root@0.0.0.0:26257/%s?sslmode=disable"

//go:embed schema.sql
var schemaFile embed.FS

func Migrate(ctx context.Context, connString string) error {
	_, err := exec.LookPath("atlas")
	if err != nil {
		return err
	}

	schemaSQL, err := schemaFile.ReadFile("schema.sql")
	if err != nil {
		return fmt.Errorf("failed to read embedded schema file: %w", err)
	}

	// Write the schemaSQL to a temporary file
	tmpFile, err := os.CreateTemp("", "schema-*.sql")
	if err != nil {
		return fmt.Errorf("failed to create temporary file: %w", err)
	}
	defer tmpFile.Close()
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.Write(schemaSQL); err != nil {
		return fmt.Errorf("failed to write to temporary file: %w", err)
	}

	args := []string{
		"schema",
		"apply",
		"--auto-approve",
		"--dev-url",
		connString,
		"-u",
		connString,
		"-f",
		tmpFile.Name(),
	}

	out, err := exec.CommandContext(ctx, "atlas", args...).CombinedOutput()
	if err != nil {
		return err
	}

	log.Info("atlas output", zap.String("out", fmt.Sprintf("out: %s", out)))

	return nil
}

func MigrateTestDB(t *testing.T, ctx context.Context) (string, *sqlx.DB) {
	_, moduleDir, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("Could not get directory path for utils/testing.")
	}

	dbName := "pacioli_test_" + strings.Replace(uuid.NewString(), "-", "", 4)
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

	cmd := exec.CommandContext(ctx, "atlas", "schema", "apply", "--auto-approve", "-u", connString, "-f", filepath.Join(moduleDir, "../schema.hcl"))
	if err = cmd.Run(); err != nil {
		t.Fatal(err)
	}

	return connString, db
}
