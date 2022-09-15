package migrations_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"gitlab.com/fynbos/backend/agreements/migrations"
	test_utils "gitlab.com/fynbos/backend/utils"
)

func TestProdAgreements(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	db := test_utils.MigrateCockroachDB(t, ctx)
	if err := migrations.MigrateFromMarkdowns(ctx, db, "assets/testing"); err != nil {
		t.Fatal(err)
	}
}

func TestMigrationIdempotency(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	db := test_utils.MigrateCockroachDB(t, ctx)
	if err := migrations.MigrateFromMarkdowns(ctx, db, "assets/testing"); err != nil {
		t.Fatal(err)
	}
	err := migrations.MigrateFromMarkdowns(ctx, db, "assets/testing")

	assert.NoError(t, err)
}
