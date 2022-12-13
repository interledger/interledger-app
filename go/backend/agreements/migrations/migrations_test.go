package migrations_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"gitlab.com/fynbos/backend/agreements/migrations"
	"gitlab.com/fynbos/backend/db"
)

func TestProdAgreements(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	db := db.MigrateTestDB(t, ctx)
	if err := migrations.MigrateFromMarkdowns(ctx, db, "assets/testing"); err != nil {
		t.Fatal(err)
	}
}

func TestMigrationIdempotency(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	db := db.MigrateTestDB(t, ctx)
	if err := migrations.MigrateFromMarkdowns(ctx, db, "assets/testing"); err != nil {
		t.Fatal(err)
	}
	err := migrations.MigrateFromMarkdowns(ctx, db, "assets/testing")

	assert.NoError(t, err)
}
