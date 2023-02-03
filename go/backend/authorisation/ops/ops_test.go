package ops_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gitlab.com/fynbos/backend/authorisation"
	"gitlab.com/fynbos/backend/authorisation/ops"
	"gitlab.com/fynbos/backend/db"
)

func TestCreateGrant(t *testing.T) {

	ctx := context.Background()
	b := ops.NewTestBackends(t, db.MigrateTestDB(t, ctx))

	err := ops.CreateClient(ctx, b, "https://fynbos.me/alice")
	require.NoError(t, err)

	g, err := ops.CreateGrant(ctx, b, authorisation.GrantRequest{
		AccessToken: []authorisation.AccessTokenReq{{
			Access: []authorisation.Access{{
				Type:      "incoming-payment",
				Actions:   []string{"create", "read"},
				Locations: []string{"https://fynbos.me/bobby"},
				Datatypes: []string{""},
			}},
			Label: "TestAccess1",
		}},
		Client: authorisation.ClientReq{
			Display: authorisation.Display{Name: "Not Used", URI: "https://fynbos.me/alice"},
			Key: authorisation.Key{
				Proof: "TODO",
			},
		},
	})
	require.NoError(t, err)

	fmt.Println(g)
}

func TestCreateAndListClientKeys(t *testing.T) {
	ctx := context.Background()
	b := ops.NewTestBackends(t, db.MigrateTestDB(t, ctx))

	err := ops.CreateClient(ctx, b, "https://fynbos.me/alice")
	require.NoError(t, err)

	key := authorisation.Jwk{
		Kty: "OKP",
		Kid: "test-key-123",
		Crv: "ed25519",
		Alg: "edDSA",
		Use: "sign",
		X:   "encoded key",
	}
	err = ops.CreateClientPublicKey(ctx, b, "https://fynbos.me/alice", key)
	require.NoError(t, err)

	getKey, err := ops.GetClientPublicKey(ctx, b, "https://fynbos.me/alice", "test-key-123")
	require.NoError(t, err)
	assert.Equal(t, key.Kty, getKey.Kty)
	assert.Equal(t, key.Alg, getKey.Alg)
	assert.Equal(t, key.Crv, getKey.Crv)
	assert.Equal(t, key.Kid, getKey.Kid)
	assert.Equal(t, key.Use, getKey.Use)
	assert.Equal(t, key.X, getKey.X)

	keys, err := ops.ListKeys(ctx, b, "https://fynbos.me/alice")
	require.NoError(t, err)
	require.Len(t, keys, 1)
	assert.Equal(t, keys[0].Kty, getKey.Kty)
	assert.Equal(t, keys[0].Alg, getKey.Alg)
	assert.Equal(t, keys[0].Crv, getKey.Crv)
	assert.Equal(t, keys[0].Kid, getKey.Kid)
	assert.Equal(t, keys[0].Use, getKey.Use)
	assert.Equal(t, keys[0].X, getKey.X)
}
