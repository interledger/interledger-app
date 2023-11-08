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

	_ "github.com/golang-migrate/migrate/v4/database/cockroachdb"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"gitlab.com/fynbos/log"
	"go.uber.org/zap"
)

const testingCrdbConnectionString = "postgres://root@0.0.0.0:26257/%s?sslmode=disable"

func Migrate(ctx context.Context, connString string) (*sqlx.DB, error) {
	_, moduleDir, _, ok := runtime.Caller(0)
	if !ok {
		return nil, fmt.Errorf("Could not get directory path for utils/testing.")
	}

	_, err := exec.LookPath("atlas")
	if err != nil {
		return nil, err
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
		filepath.Join(moduleDir, "../schema.sql"),
	}

	out, err := exec.CommandContext(ctx, "atlas", args...).CombinedOutput()
	if err != nil {
		return nil, err
	}

	log.Info("atlas output", zap.String("out", fmt.Sprintf("out: %s", out)))

	// TODO:THe things where we connect to the DB and run the migration.
	return nil, nil
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
