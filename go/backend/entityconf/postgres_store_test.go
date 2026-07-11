package entityconf_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/interledger/interledger-app/go/backend/db"
	"github.com/interledger/interledger-app/go/backend/entityconf"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPostgresStore(t *testing.T) {
	runStoreContractTests(t, func(t *testing.T) entityconf.Store {
		t.Helper()
		conn := db.MigrateTestDB(t, context.Background(), "")
		return entityconf.NewPostgresStore(conn)
	})
}

// This test isn't part of the shared contract (it's specific to what makes
// the Postgres Store different from the in-memory one): confs really
// persist as rows, so two Store values backed by the same connection see
// each other's writes — unlike two independent in-memory stores.
func TestPostgresStore_PersistsAcrossStoreInstances(t *testing.T) {
	ctx := context.Background()
	conn := db.MigrateTestDB(t, ctx, "")

	walletID := uuid.NewString()

	writer := entityconf.NewPostgresStore(conn)
	require.NoError(t, writer.SyncDefinitions(ctx, []entityconf.Definition{
		{Key: "wallet.persisted", EntityType: testWallet, Type: entityconf.TypeBool, CodeDefault: false},
	}))
	require.NoError(t, writer.SetValue(ctx, testWallet, walletID, "wallet.persisted", true))

	reader := entityconf.NewPostgresStore(conn)
	values, err := reader.ResolveAll(ctx, testWallet, walletID, []string{"wallet.persisted"})
	require.NoError(t, err)
	assert.Equal(t, map[string]any{"wallet.persisted": true}, values)
}

// A stored definition with an unrecognized value type can't be decoded back
// out of jsonb — StoredDefinition/StoredDefinitions must surface that as an
// error rather than panic or silently drop the row. Postgres-specific: the
// in-memory Store never serializes values, so it has nothing analogous to
// decode.
func TestPostgresStore_StoredDefinition_ErrorsOnUnrecognizedValueType(t *testing.T) {
	ctx := context.Background()
	conn := db.MigrateTestDB(t, ctx, "")
	store := entityconf.NewPostgresStore(conn)

	require.NoError(t, store.SyncDefinitions(ctx, []entityconf.Definition{
		{Key: "wallet.weird", EntityType: testWallet, Type: entityconf.ValueType("unknown"), CodeDefault: nil},
	}))

	_, err := store.StoredDefinition(ctx, "wallet.weird")
	require.Error(t, err)

	_, err = store.StoredDefinitions(ctx)
	require.Error(t, err)
}
