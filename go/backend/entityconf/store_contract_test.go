package entityconf_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/interledger/interledger-app/go/backend/entityconf"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func sampleDefs() []entityconf.Definition {
	return []entityconf.Definition{
		{Key: "wallet.a", EntityType: testWallet, Type: entityconf.TypeBool, DisplayName: "A", Description: "a", CodeDefault: true},
		{Key: "wallet.b", EntityType: testWallet, Type: entityconf.TypeInt, DisplayName: "B", Description: "b", CodeDefault: 5},
	}
}

// runStoreContractTests exercises the full Store contract against a fresh
// store returned by newStore for each subtest. Every Store implementation
// (in-memory, Postgres, ...) must behave identically here — this is the one
// place that behavior is specified and verified for all of them at once.
//
// Entity IDs use real UUIDs throughout: the Postgres-backed Store's
// entity_id column is uuid-typed, matching every real entity type in this
// app (wallets, and any future entity type) — the in-memory Store treats
// entity IDs as opaque strings, so this is equally valid there.
func runStoreContractTests(t *testing.T, newStore func(t *testing.T) entityconf.Store) {
	t.Helper()

	t.Run("SyncDefinitions seeds effective default once and preserves admin edits across resync", func(t *testing.T) {
		ctx := context.Background()
		store := newStore(t)
		require.NoError(t, store.SyncDefinitions(ctx, sampleDefs()))

		def, err := store.StoredDefinition(ctx, "wallet.a")
		require.NoError(t, err)
		assert.Equal(t, true, def.EffectiveDefault)
		assert.Nil(t, def.DeprecatedAt)

		// An admin changes the effective default...
		require.NoError(t, store.SetEffectiveDefault(ctx, "wallet.a", false))

		// ...a later sync must not clobber it, even though CodeDefault is
		// re-synced every time.
		require.NoError(t, store.SyncDefinitions(ctx, sampleDefs()))

		def, err = store.StoredDefinition(ctx, "wallet.a")
		require.NoError(t, err)
		assert.Equal(t, false, def.EffectiveDefault, "effective default must survive a re-sync")
		assert.Equal(t, true, def.CodeDefault, "code default must still refresh from the registry")
	})

	t.Run("SyncDefinitions deprecates and un-deprecates keys", func(t *testing.T) {
		ctx := context.Background()
		store := newStore(t)
		defs := sampleDefs()
		require.NoError(t, store.SyncDefinitions(ctx, defs))

		// Second sync omits wallet.b.
		require.NoError(t, store.SyncDefinitions(ctx, defs[:1]))

		def, err := store.StoredDefinition(ctx, "wallet.b")
		require.NoError(t, err)
		assert.NotNil(t, def.DeprecatedAt)

		// Re-registering it un-deprecates.
		require.NoError(t, store.SyncDefinitions(ctx, defs))
		def, err = store.StoredDefinition(ctx, "wallet.b")
		require.NoError(t, err)
		assert.Nil(t, def.DeprecatedAt)
	})

	t.Run("StoredDefinition not found", func(t *testing.T) {
		store := newStore(t)
		_, err := store.StoredDefinition(context.Background(), "nope")
		assert.ErrorIs(t, err, entityconf.ErrDefinitionNotFound)
	})

	t.Run("StoredDefinitions sorted by key", func(t *testing.T) {
		ctx := context.Background()
		store := newStore(t)
		require.NoError(t, store.SyncDefinitions(ctx, sampleDefs()))

		defs, err := store.StoredDefinitions(ctx)
		require.NoError(t, err)
		require.Len(t, defs, 2)
		assert.Equal(t, "wallet.a", defs[0].Key)
		assert.Equal(t, "wallet.b", defs[1].Key)
	})

	t.Run("SetEffectiveDefault errors", func(t *testing.T) {
		ctx := context.Background()
		store := newStore(t)
		require.NoError(t, store.SyncDefinitions(ctx, sampleDefs()))

		err := store.SetEffectiveDefault(ctx, "nope", true)
		assert.ErrorIs(t, err, entityconf.ErrDefinitionNotFound)

		err = store.SetEffectiveDefault(ctx, "wallet.a", "not-a-bool")
		assert.ErrorIs(t, err, entityconf.ErrTypeMismatch)
	})

	t.Run("ResolveAll", func(t *testing.T) {
		ctx := context.Background()
		store := newStore(t)
		require.NoError(t, store.SyncDefinitions(ctx, sampleDefs()))

		w1, w2 := uuid.NewString(), uuid.NewString()

		values, err := store.ResolveAll(ctx, testWallet, w1, []string{"wallet.a", "wallet.b", "unknown.key"})
		require.NoError(t, err)
		assert.Equal(t, map[string]any{"wallet.a": true, "wallet.b": 5}, values)

		require.NoError(t, store.SetValue(ctx, testWallet, w1, "wallet.a", false))
		values, err = store.ResolveAll(ctx, testWallet, w1, []string{"wallet.a", "wallet.b"})
		require.NoError(t, err)
		assert.Equal(t, map[string]any{"wallet.a": false, "wallet.b": 5}, values)

		// A different entity instance is unaffected.
		values, err = store.ResolveAll(ctx, testWallet, w2, []string{"wallet.a"})
		require.NoError(t, err)
		assert.Equal(t, map[string]any{"wallet.a": true}, values)
	})

	t.Run("SetValue errors", func(t *testing.T) {
		ctx := context.Background()
		store := newStore(t)
		require.NoError(t, store.SyncDefinitions(ctx, sampleDefs()))

		w1, s1 := uuid.NewString(), uuid.NewString()

		err := store.SetValue(ctx, testWallet, w1, "nope", true)
		assert.ErrorIs(t, err, entityconf.ErrDefinitionNotFound)

		err = store.SetValue(ctx, testBattleShip, s1, "wallet.a", true)
		assert.ErrorIs(t, err, entityconf.ErrEntityTypeMismatch)

		err = store.SetValue(ctx, testWallet, w1, "wallet.a", "not-a-bool")
		assert.ErrorIs(t, err, entityconf.ErrTypeMismatch)
	})

	t.Run("ClearValue is idempotent and reverts to default", func(t *testing.T) {
		ctx := context.Background()
		store := newStore(t)
		require.NoError(t, store.SyncDefinitions(ctx, sampleDefs()))

		w1 := uuid.NewString()

		// Clearing something never set must not error.
		require.NoError(t, store.ClearValue(ctx, testWallet, w1, "wallet.a"))

		require.NoError(t, store.SetValue(ctx, testWallet, w1, "wallet.a", false))
		require.NoError(t, store.ClearValue(ctx, testWallet, w1, "wallet.a"))

		values, err := store.ResolveAll(ctx, testWallet, w1, []string{"wallet.a"})
		require.NoError(t, err)
		assert.Equal(t, map[string]any{"wallet.a": true}, values)
	})

	t.Run("SetValue accepts each supported value type", func(t *testing.T) {
		ctx := context.Background()
		store := newStore(t)
		require.NoError(t, store.SyncDefinitions(ctx, []entityconf.Definition{
			{Key: "wallet.bool", EntityType: testWallet, Type: entityconf.TypeBool, CodeDefault: false},
			{Key: "wallet.int", EntityType: testWallet, Type: entityconf.TypeInt, CodeDefault: 0},
			{Key: "wallet.string", EntityType: testWallet, Type: entityconf.TypeString, CodeDefault: ""},
		}))

		w1 := uuid.NewString()

		require.NoError(t, store.SetValue(ctx, testWallet, w1, "wallet.bool", true))
		require.NoError(t, store.SetValue(ctx, testWallet, w1, "wallet.int", 42))
		require.NoError(t, store.SetValue(ctx, testWallet, w1, "wallet.string", "hello"))

		values, err := store.ResolveAll(ctx, testWallet, w1, []string{"wallet.bool", "wallet.int", "wallet.string"})
		require.NoError(t, err)
		assert.Equal(t, map[string]any{"wallet.bool": true, "wallet.int": 42, "wallet.string": "hello"}, values)

		assert.ErrorIs(t, store.SetValue(ctx, testWallet, w1, "wallet.int", "not-an-int"), entityconf.ErrTypeMismatch)
		assert.ErrorIs(t, store.SetValue(ctx, testWallet, w1, "wallet.string", 42), entityconf.ErrTypeMismatch)
	})

	t.Run("rejects values for an unrecognized value type", func(t *testing.T) {
		// A defensive-programming case: nothing in the public API can
		// produce a Definition with a Type outside
		// TypeBool/TypeInt/TypeString today, but Store implementations
		// must not silently accept a value for one if a caller ever
		// manages to synthesize one — every value should be rejected as a
		// mismatch.
		ctx := context.Background()
		store := newStore(t)
		require.NoError(t, store.SyncDefinitions(ctx, []entityconf.Definition{
			{Key: "wallet.weird", EntityType: testWallet, Type: entityconf.ValueType("unknown"), CodeDefault: nil},
		}))

		err := store.SetEffectiveDefault(ctx, "wallet.weird", "anything")
		assert.ErrorIs(t, err, entityconf.ErrTypeMismatch)

		err = store.SetValue(ctx, testWallet, uuid.NewString(), "wallet.weird", "anything")
		assert.ErrorIs(t, err, entityconf.ErrTypeMismatch)
	})

	t.Run("concurrent access", func(t *testing.T) {
		ctx := context.Background()
		store := newStore(t)
		require.NoError(t, store.SyncDefinitions(ctx, sampleDefs()))

		w1 := uuid.NewString()

		const workers = 20
		done := make(chan struct{}, workers)
		for i := range workers {
			go func(i int) {
				defer func() { done <- struct{}{} }()
				_ = store.SetValue(ctx, testWallet, w1, "wallet.a", i%2 == 0)
				_, _ = store.ResolveAll(ctx, testWallet, w1, []string{"wallet.a", "wallet.b"})
			}(i)
		}
		for range workers {
			<-done
		}
	})
}
