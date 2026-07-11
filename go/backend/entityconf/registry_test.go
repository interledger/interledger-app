package entityconf_test

import (
	"testing"

	"github.com/interledger/interledger-app/go/backend/entityconf"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRegistry_Register_Success(t *testing.T) {
	t.Parallel()

	r := entityconf.NewRegistry()
	require.NoError(t, r.Register(testWallet, walletConfsFixture{}))

	defs := r.DefinitionsFor(testWallet)
	require.Len(t, defs, 4)

	byKey := map[string]entityconf.Definition{}
	for _, d := range defs {
		byKey[d.Key] = d
	}

	sendDef, ok := byKey["wallet.send_enabled"]
	require.True(t, ok)
	assert.Equal(t, testWallet, sendDef.EntityType)
	assert.Equal(t, entityconf.TypeBool, sendDef.Type)
	assert.Equal(t, "Send Payments", sendDef.DisplayName)
	assert.Equal(t, "Allows sending", sendDef.Description)
	assert.Equal(t, true, sendDef.CodeDefault)

	receiveDef, ok := byKey["wallet.receive_enabled"]
	require.True(t, ok)
	assert.Equal(t, false, receiveDef.CodeDefault)

	limitDef, ok := byKey["wallet.max_daily_limit"]
	require.True(t, ok)
	assert.Equal(t, entityconf.TypeInt, limitDef.Type)
	assert.Equal(t, 1000, limitDef.CodeDefault)

	nameDef, ok := byKey["wallet.nickname"]
	require.True(t, ok)
	assert.Equal(t, entityconf.TypeString, nameDef.Type)
	assert.Equal(t, "", nameDef.CodeDefault)

	// The "-"-tagged and unexported fields must never become confs.
	for _, d := range defs {
		assert.NotContains(t, d.Key, "skipped")
		assert.NotContains(t, d.Key, "internal")
	}
}

func TestRegistry_Register_RejectsNonStruct(t *testing.T) {
	t.Parallel()

	r := entityconf.NewRegistry()

	err := r.Register(testWallet, &walletConfsFixture{})
	assert.ErrorIs(t, err, entityconf.ErrNotAStruct)

	err = r.Register(testWallet, 42)
	assert.ErrorIs(t, err, entityconf.ErrNotAStruct)
}

func TestRegistry_Register_RejectsDoubleRegistrationOfSameType(t *testing.T) {
	t.Parallel()

	r := entityconf.NewRegistry()
	require.NoError(t, r.Register(testWallet, walletConfsFixture{}))

	err := r.Register(testWallet, walletConfsFixture{})
	assert.ErrorIs(t, err, entityconf.ErrTypeAlreadyRegistered)
}

func TestRegistry_Register_RejectsMissingConfTag(t *testing.T) {
	t.Parallel()

	type missingTagFixture struct {
		Foo bool
	}

	r := entityconf.NewRegistry()
	err := r.Register(testWallet, missingTagFixture{})
	assert.ErrorIs(t, err, entityconf.ErrMissingConfTag)
	assert.Empty(t, r.Definitions())
}

func TestRegistry_Register_RejectsEmptyKey(t *testing.T) {
	t.Parallel()

	type emptyKeyFixture struct {
		Foo bool `conf:""`
	}

	r := entityconf.NewRegistry()
	err := r.Register(testWallet, emptyKeyFixture{})
	assert.ErrorIs(t, err, entityconf.ErrEmptyConfKey)
}

func TestRegistry_Register_RejectsBadPrefix(t *testing.T) {
	t.Parallel()

	type badPrefixFixture struct {
		Foo bool `conf:"other.thing" default:"true"`
	}

	r := entityconf.NewRegistry()
	err := r.Register(testWallet, badPrefixFixture{})
	assert.ErrorIs(t, err, entityconf.ErrInvalidKeyPrefix)
}

func TestRegistry_Register_RejectsKeyThatIsJustThePrefix(t *testing.T) {
	t.Parallel()

	type emptyAfterPrefixFixture struct {
		Foo bool `conf:"wallet." default:"true"`
	}

	r := entityconf.NewRegistry()
	err := r.Register(testWallet, emptyAfterPrefixFixture{})
	assert.ErrorIs(t, err, entityconf.ErrInvalidKeyPrefix)
}

func TestRegistry_Register_RejectsDuplicateKeyAcrossTypes(t *testing.T) {
	t.Parallel()

	type dupeFixtureA struct {
		Foo bool `conf:"wallet.dupe_x" default:"true"`
	}
	type dupeFixtureB struct {
		Foo bool `conf:"wallet.dupe_x" default:"false"`
	}

	r := entityconf.NewRegistry()
	require.NoError(t, r.Register(testWallet, dupeFixtureA{}))

	err := r.Register(testWallet, dupeFixtureB{})
	assert.ErrorIs(t, err, entityconf.ErrDuplicateKey)
}

func TestRegistry_Register_RejectsDuplicateKeyWithinStruct(t *testing.T) {
	t.Parallel()

	type dupeWithinFixture struct {
		Foo bool `conf:"wallet.dupe_y" default:"true"`
		Bar bool `conf:"wallet.dupe_y" default:"false"`
	}

	r := entityconf.NewRegistry()
	err := r.Register(testWallet, dupeWithinFixture{})
	assert.ErrorIs(t, err, entityconf.ErrDuplicateKey)
	assert.Empty(t, r.Definitions(), "a failed registration must not partially commit")
}

func TestRegistry_Register_RejectsUnsupportedFieldKind(t *testing.T) {
	t.Parallel()

	type unsupportedKindFixture struct {
		Foo float64 `conf:"wallet.foo" default:"1.5"`
	}

	r := entityconf.NewRegistry()
	err := r.Register(testWallet, unsupportedKindFixture{})
	assert.ErrorIs(t, err, entityconf.ErrUnsupportedFieldKind)
}

func TestRegistry_Register_RejectsUnparseableDefault(t *testing.T) {
	t.Parallel()

	t.Run("bool", func(t *testing.T) {
		t.Parallel()
		type badBoolDefaultFixture struct {
			Foo bool `conf:"wallet.foo" default:"notabool"`
		}
		r := entityconf.NewRegistry()
		err := r.Register(testWallet, badBoolDefaultFixture{})
		assert.ErrorIs(t, err, entityconf.ErrInvalidDefaultTag)
	})

	t.Run("int", func(t *testing.T) {
		t.Parallel()
		type badIntDefaultFixture struct {
			Foo int `conf:"wallet.foo" default:"notanumber"`
		}
		r := entityconf.NewRegistry()
		err := r.Register(testWallet, badIntDefaultFixture{})
		assert.ErrorIs(t, err, entityconf.ErrInvalidDefaultTag)
	})
}

func TestRegistry_MustRegister_PanicsOnError(t *testing.T) {
	t.Parallel()

	type missingTagFixture struct {
		Foo bool
	}

	r := entityconf.NewRegistry()
	assert.Panics(t, func() {
		r.MustRegister(testWallet, missingTagFixture{})
	})
}

func TestRegistry_MustRegister_Succeeds(t *testing.T) {
	t.Parallel()

	r := entityconf.NewRegistry()
	assert.NotPanics(t, func() {
		r.MustRegister(testWallet, walletConfsFixture{})
	})
	assert.Len(t, r.Definitions(), 4)
}

func TestRegistry_Definitions_And_DefinitionsFor(t *testing.T) {
	t.Parallel()

	r := entityconf.NewRegistry()
	require.NoError(t, r.Register(testWallet, walletConfsFixture{}))
	require.NoError(t, r.Register(testBattleShip, shipConfs{}))

	all := r.Definitions()
	assert.Len(t, all, 8)

	walletOnly := r.DefinitionsFor(testWallet)
	assert.Len(t, walletOnly, 4)
	for _, d := range walletOnly {
		assert.Equal(t, testWallet, d.EntityType)
	}

	shipOnly := r.DefinitionsFor(testBattleShip)
	assert.Len(t, shipOnly, 4)
	for _, d := range shipOnly {
		assert.Equal(t, testBattleShip, d.EntityType)
	}
}
