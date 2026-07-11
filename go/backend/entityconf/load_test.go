package entityconf_test

import (
	"context"
	"errors"
	"testing"

	"github.com/interledger/interledger-app/go/backend/entityconf"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newLoadedRegistry(t *testing.T) (*entityconf.Registry, entityconf.Store) {
	t.Helper()

	r := entityconf.NewRegistry()
	require.NoError(t, r.Register(testBattleShip, shipConfs{}))

	store := entityconf.NewInMemoryStore()
	require.NoError(t, store.SyncDefinitions(context.Background(), r.Definitions()))

	return r, store
}

func TestRegistry_Load_DefaultsWhenNoOverride(t *testing.T) {
	t.Parallel()

	r, store := newLoadedRegistry(t)

	var confs shipConfs
	require.NoError(t, r.Load(context.Background(), store, "ship-1", &confs))

	assert.Equal(t, "USS Unnamed", confs.Name)
	assert.Equal(t, 250, confs.Length)
	assert.True(t, confs.HasFrontTurret)
	assert.True(t, confs.HasBackTurret)
}

func TestRegistry_Load_ReflectsPerEntityOverride(t *testing.T) {
	t.Parallel()

	r, store := newLoadedRegistry(t)
	ctx := context.Background()

	require.NoError(t, store.SetValue(ctx, testBattleShip, "ship-1", "battleship.has_back_turret", false))

	var confs shipConfs
	require.NoError(t, r.Load(ctx, store, "ship-1", &confs))
	assert.False(t, confs.HasBackTurret)

	// A different instance of the same entity type is unaffected.
	var other shipConfs
	require.NoError(t, r.Load(ctx, store, "ship-2", &other))
	assert.True(t, other.HasBackTurret)
}

func TestRegistry_Load_ClearValueRevertsToDefault(t *testing.T) {
	t.Parallel()

	r, store := newLoadedRegistry(t)
	ctx := context.Background()

	require.NoError(t, store.SetValue(ctx, testBattleShip, "ship-1", "battleship.has_back_turret", false))
	require.NoError(t, store.ClearValue(ctx, testBattleShip, "ship-1", "battleship.has_back_turret"))

	var confs shipConfs
	require.NoError(t, r.Load(ctx, store, "ship-1", &confs))
	assert.True(t, confs.HasBackTurret)
}

// This test documents plan.md §8 item 1: changing the effective default is
// retroactive for every entity that has no explicit override — including
// ones that already existed before the default changed.
func TestRegistry_Load_EffectiveDefaultChangeIsRetroactive(t *testing.T) {
	t.Parallel()

	r, store := newLoadedRegistry(t)
	ctx := context.Background()

	// ship-1 never had an override.
	var before shipConfs
	require.NoError(t, r.Load(ctx, store, "ship-1", &before))
	require.True(t, before.HasBackTurret)

	require.NoError(t, store.SetEffectiveDefault(ctx, "battleship.has_back_turret", false))

	var after shipConfs
	require.NoError(t, r.Load(ctx, store, "ship-1", &after))
	assert.False(t, after.HasBackTurret)
}

func TestRegistry_Load_RejectsNonPointer(t *testing.T) {
	t.Parallel()

	r, store := newLoadedRegistry(t)
	err := r.Load(context.Background(), store, "ship-1", shipConfs{})
	assert.ErrorIs(t, err, entityconf.ErrNotAPointer)
}

func TestRegistry_Load_RejectsNilPointer(t *testing.T) {
	t.Parallel()

	r, store := newLoadedRegistry(t)
	var confs *shipConfs
	err := r.Load(context.Background(), store, "ship-1", confs)
	assert.ErrorIs(t, err, entityconf.ErrNotAPointer)
}

func TestRegistry_Load_RejectsUnregisteredType(t *testing.T) {
	t.Parallel()

	type unregisteredConfs struct {
		Foo bool `conf:"unregistered.foo" default:"true"`
	}

	r, store := newLoadedRegistry(t)
	var confs unregisteredConfs
	err := r.Load(context.Background(), store, "x", &confs)
	assert.ErrorIs(t, err, entityconf.ErrTypeNotRegistered)
}

func TestRegistry_Load_ErrorsWhenNotSynced(t *testing.T) {
	t.Parallel()

	r := entityconf.NewRegistry()
	require.NoError(t, r.Register(testBattleShip, shipConfs{}))
	store := entityconf.NewInMemoryStore() // note: SyncDefinitions never called

	var confs shipConfs
	err := r.Load(context.Background(), store, "ship-1", &confs)
	assert.ErrorIs(t, err, entityconf.ErrDefinitionNotFound)
}

// erroringStore wraps a real Store but forces ResolveAll to fail, to verify
// Load propagates the underlying store's errors rather than swallowing them.
type erroringStore struct {
	entityconf.Store
}

func (erroringStore) ResolveAll(context.Context, entityconf.EntityType, string, []string) (map[string]any, error) {
	return nil, errors.New("boom")
}

func TestRegistry_Load_PropagatesStoreError(t *testing.T) {
	t.Parallel()

	r, store := newLoadedRegistry(t)
	var confs shipConfs
	err := r.Load(context.Background(), erroringStore{Store: store}, "ship-1", &confs)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "boom")
}

// mismatchedTypeStore wraps a real Store but forces ResolveAll to return a
// value of the wrong Go type for one key, to verify Load defends against a
// misbehaving Store rather than panicking via reflect.Value.Set.
type mismatchedTypeStore struct {
	entityconf.Store
}

func (s mismatchedTypeStore) ResolveAll(ctx context.Context, entityType entityconf.EntityType, entityID string, keys []string) (map[string]any, error) {
	values, err := s.Store.ResolveAll(ctx, entityType, entityID, keys)
	if err != nil {
		return nil, err
	}
	values["battleship.length"] = "not-an-int" // Length is declared as int
	return values, nil
}

func TestRegistry_Load_RejectsMismatchedValueTypeFromStore(t *testing.T) {
	t.Parallel()

	r, store := newLoadedRegistry(t)
	var confs shipConfs
	err := r.Load(context.Background(), mismatchedTypeStore{Store: store}, "ship-1", &confs)
	assert.ErrorIs(t, err, entityconf.ErrTypeMismatch)
}
