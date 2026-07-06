package migrations_test

import (
	"context"
	"testing"

	"github.com/interledger/interledger-app/go/backend/agreements/migrations"
	"github.com/interledger/interledger-app/go/backend/db"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestProdAgreements(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	db := db.MigrateTestDB(t, ctx)
	if err := migrations.MigrateFromMarkdowns(ctx, db, "assets/testing"); err != nil {
		t.Fatal(err)
	}

	var gitFilePath string
	err := db.GetContext(ctx, &gitFilePath, "SELECT git_file_path FROM agreements WHERE id = $1 LIMIT 1", "privacy_policy-2.0.0")
	require.NoError(t, err)
	assert.Contains(t, gitFilePath, "migrations/assets/testing", "git_file_path should reference the agreement file path in the repo")
}

func TestMigrationIdempotency(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	db1 := db.MigrateTestDB(t, ctx)
	if err := migrations.MigrateFromMarkdowns(ctx, db1, "assets/testing"); err != nil {
		t.Fatal(err)
	}
	// Create a new test database for the second migration run to ensure clean state
	db2 := db.MigrateTestDB(t, ctx)
	err := migrations.MigrateFromMarkdowns(ctx, db2, "assets/testing")

	assert.NoError(t, err)
}
