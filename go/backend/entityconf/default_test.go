package entityconf_test

import (
	"context"
	"testing"

	"github.com/interledger/interledger-app/go/backend/entityconf"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// These tests exercise the package-level default registry (MustRegister,
// Load, Definitions, DefinitionsFor). They deliberately use entity types and
// keys not used anywhere else in this test package, since — unlike the
// Registry-based tests — they share global state with every other test in
// this binary.

type pkgLevelConfs struct {
	Enabled bool `conf:"pkgleveltest.enabled" default:"true" display:"Enabled" desc:"desc"`
}

const entityPkgLevelTest entityconf.EntityType = "pkgleveltest"

func TestPackageLevel_MustRegisterAndLoad(t *testing.T) {
	entityconf.MustRegister(entityPkgLevelTest, pkgLevelConfs{})

	defs := entityconf.DefinitionsFor(entityPkgLevelTest)
	require.Len(t, defs, 1)
	assert.Equal(t, "pkgleveltest.enabled", defs[0].Key)

	all := entityconf.Definitions()
	assert.Contains(t, all, defs[0])

	store := entityconf.NewInMemoryStore()
	require.NoError(t, store.SyncDefinitions(context.Background(), defs))

	var confs pkgLevelConfs
	require.NoError(t, entityconf.Load(context.Background(), store, "x", &confs))
	assert.True(t, confs.Enabled)
}

func TestPackageLevel_MustRegister_PanicsOnInvalidTag(t *testing.T) {
	type badConfs struct {
		Foo bool `conf:"badpkgleveltest.foo" default:"notabool"`
	}

	assert.Panics(t, func() {
		entityconf.MustRegister(entityconf.EntityType("badpkgleveltest"), badConfs{})
	})
}
