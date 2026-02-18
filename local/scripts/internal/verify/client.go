package verify

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/joho/godotenv"
)

// Client loads configuration and manages email verification for users.
type Client struct {
	kratosDSN string
}

// NewClient constructs a client using environment variables with sensible defaults.
//
// Environment variables:
//
//	KRATOS_DATABASE_URL - DSN for the Kratos Postgres database (default: postgres://postgres:postgres@localhost:5432/kratos?sslmode=disable)
func NewClient() *Client {
	envPath := filepath.Join("..", ".env")
	_ = godotenv.Load(envPath)

	getEnv := func(key, fallback string) string {
		if val := os.Getenv(key); val != "" {
			return val
		}
		return fallback
	}

	return &Client{
		kratosDSN: getEnv("KRATOS_DATABASE_URL", "postgres://postgres:postgres@localhost:5432/kratos?sslmode=disable"),
	}
}

// VerifyEmail directly updates the database to mark a user's email as verified.
// This bypasses the email verification flow and is intended for local development use.
func (c *Client) VerifyEmail(ctx context.Context, email string) error {
	db, err := sql.Open("pgx", c.kratosDSN)
	if err != nil {
		return fmt.Errorf("open kratos database: %w", err)
	}
	defer db.Close()

	// First check if the user exists
	var exists bool
	err = db.QueryRowContext(ctx, "SELECT EXISTS(SELECT 1 FROM identity_verifiable_addresses WHERE value = $1)", email).Scan(&exists)
	if err != nil {
		return fmt.Errorf("check if user exists: %w", err)
	}
	if !exists {
		return fmt.Errorf("no user found with email: %s", email)
	}

	// Update the verification status
	const updateQuery = `
UPDATE identity_verifiable_addresses 
SET verified = true, 
    status = 'completed',
    updated_at = $1
WHERE value = $2`

	result, err := db.ExecContext(ctx, updateQuery, time.Now(), email)
	if err != nil {
		return fmt.Errorf("update verification status: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("check rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("no rows updated for email: %s (user may not exist)", email)
	}

	return nil
}
