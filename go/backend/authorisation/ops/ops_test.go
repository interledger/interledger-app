package ops_test

import (
	"context"
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
	clientPaymentPointer := "https://fynbos.me/alice"
	resourceOwnerPaymentPointer := "https://fynbos.me/bobby"
	_, err := ops.CreateClient(ctx, b, clientPaymentPointer)
	require.NoError(t, err)

	g, err := ops.CreateGrant(ctx, b, authorisation.GrantRequest{
		AccessToken: []authorisation.AccessTokenReq{{
			Access: []authorisation.Access{
				{
					Type:       "incoming-payment",
					Actions:    []string{"write", "read"},
					Identifier: resourceOwnerPaymentPointer,
				},
				{
					Type:       "outgoing-payment",
					Actions:    []string{"write", "read"},
					Identifier: resourceOwnerPaymentPointer,
				}},
			Label: "TestAccess1",
		}},
		Client: clientPaymentPointer,
	})
	require.NoError(t, err)

	assert.Len(t, g.Tokens, 1)
	assert.Len(t, g.Tokens[0].Access, 2)
	for _, access := range g.Tokens[0].Access {
		if access.Type == "incoming-payment" {
			assert.EqualValues(t, access.Locations, []string{"https://fynbos.me/incoming"})
			assert.EqualValues(t, access.Actions, []string{"write", "read"})
			continue
		}

		if access.Type == "outgoing-payment" {
			assert.EqualValues(t, access.Locations, []string{"https://fynbos.me/outgoing"})
			assert.EqualValues(t, access.Actions, []string{"write", "read"})
			continue
		}

		t.Fatal("Only expected outgoing-payment and incoming-payment access types")
	}

	grant, err := ops.Introspect(ctx, b, g.Tokens[0].Value)
	require.NoError(t, err)
	assert.Len(t, grant.Tokens, 1)
	assert.Len(t, grant.Tokens[0].Access, 2)
	for _, access := range grant.Tokens[0].Access {
		if access.Type == "incoming-payment" {
			assert.EqualValues(t, access.Locations, []string{"https://fynbos.me/incoming"})
			assert.EqualValues(t, access.Actions, []string{"write", "read"})
			continue
		}

		if access.Type == "outgoing-payment" {
			assert.EqualValues(t, access.Locations, []string{"https://fynbos.me/outgoing"})
			assert.EqualValues(t, access.Actions, []string{"write", "read"})
			continue
		}

		t.Fatal("Only expected outgoing-payment and incoming-payment access types")
	}
}

func TestCreateAndListClientKeys(t *testing.T) {
	ctx := context.Background()
	b := ops.NewTestBackends(t, db.MigrateTestDB(t, ctx))

	_, err := ops.CreateClient(ctx, b, "https://fynbos.me/alice")
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
