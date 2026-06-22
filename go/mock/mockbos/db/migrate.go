package db

import (
	"context"
	"embed"
	"fmt"
	"github.com/interledger/interledger-app/go/log"
	"github.com/jackc/pgx/v5"
	"go.uber.org/zap"
	url2 "net/url"
	"os"
	"os/exec"
)

const tempDbName = "temp_atlas"

//go:embed schema.sql
var schemaFile embed.FS

func Migrate(ctx context.Context, connString string, conn *pgx.Conn) error {
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

	// Create temp table for dev url
	url, err := url2.Parse(connString)
	if err != nil {
		return fmt.Errorf("failed to parse database url: %w", err)
	}
	url.Path = tempDbName
	_, err = conn.Exec(ctx, "drop database if exists temp_atlas;")
	if err != nil {
		return fmt.Errorf("failed to drop database: %w", err)
	}
	_, err = conn.Exec(ctx, "create database temp_atlas;")
	if err != nil {
		return fmt.Errorf("failed to create database: %w", err)
	}

	args := []string{
		"schema",
		"apply",
		"--auto-approve",
		"--dev-url",
		url.String(),
		"-u",
		connString,
		"--to",
		fmt.Sprintf("file://%s", tmpFile.Name()),
	}

	out, err := exec.CommandContext(ctx, "atlas", args...).CombinedOutput()
	log.Info("atlas output", zap.String("out", fmt.Sprintf("out: %s", out)))
	if err != nil {
		log.Error("atlas error", zap.Error(err))
	}

	_, err = conn.Exec(ctx, "drop database if exists temp_atlas;")

	return err
}
