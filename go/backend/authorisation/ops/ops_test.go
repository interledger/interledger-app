package ops_test

import (
	"context"
	"testing"

	"gitlab.com/fynbos/backend/wallets"

	wallets_mock "gitlab.com/fynbos/backend/wallets/client/mock"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gitlab.com/fynbos/backend/authorisation"
	"gitlab.com/fynbos/backend/authorisation/ops"
	"gitlab.com/fynbos/backend/db"
)

func TestCreateClient(t *testing.T) {
	cases := []struct {
		name          string
		address       string
		addressExists bool
		err           error
	}{
		{
			name:          "success",
			address:       "https://ilp.link/tables",
			addressExists: true,
			err:           nil,
		},
		{
			name:          "does not exist",
			address:       "https://ilp.link/notehere",
			addressExists: false,
			err:           wallets.ErrNoWalletFound,
		},
	}

	ctx := context.Background()

	ctrl := gomock.NewController(t)
	wc := wallets_mock.NewMockClient(ctrl)

	b := ops.NewTestBackends(t, db.MigrateTestDB(t, ctx), wc)

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.addressExists {
				wa, err := wallets.ParseAddress(tc.address)
				require.NoError(t, err)
				wc.EXPECT().GetFromAddress(ctx, tc.address).Return(&wallets.Wallet{Addresses: []wallets.Address{wa}}, nil).Times(1)
			} else {
				wc.EXPECT().GetFromAddress(ctx, tc.address).Return(nil, wallets.ErrNoWalletFound).Times(1)
			}
			_, err := ops.CreateClient(ctx, b, tc.address)
			if tc.err != nil {
				require.ErrorIs(t, err, tc.err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestCreateGrant(t *testing.T) {
	ctx := context.Background()

	ctrl := gomock.NewController(t)
	wc := wallets_mock.NewMockClient(ctrl)

	b := ops.NewTestBackends(t, db.MigrateTestDB(t, ctx), wc)
	clientAddress, err := wallets.ParseAddress("https://ilp.link/alice")
	require.NoError(t, err)
	resourceAddress, err := wallets.ParseAddress("https://ilp.link/bobby")
	require.NoError(t, err)

	wc.EXPECT().GetFromAddress(ctx, clientAddress.String()).Return(&wallets.Wallet{Addresses: []wallets.Address{clientAddress}}, nil).AnyTimes()
	wc.EXPECT().GetFromAddress(ctx, resourceAddress.String()).Return(&wallets.Wallet{Addresses: []wallets.Address{resourceAddress}}, nil).AnyTimes()

	_, err = ops.CreateClient(ctx, b, clientAddress.String())
	require.NoError(t, err)

	g, err := ops.CreateGrant(ctx, b, authorisation.GrantRequest{
		AccessToken: []authorisation.AccessTokenReq{{
			Access: []authorisation.Access{
				{
					// To be ignored, you can only get tokens for yourself, currently
					Type:       "incoming-payment",
					Actions:    []string{"write", "read"},
					Identifier: resourceAddress.String(),
				},
				{
					// To be ignored, you can only get tokens for yourself, currently
					Type:       "outgoing-payment",
					Actions:    []string{"write", "read"},
					Identifier: resourceAddress.String(),
				},
				{
					Type:       "incoming-payment",
					Actions:    []string{"write", "read"},
					Identifier: clientAddress.String(),
				},
				{
					Type:       "outgoing-payment",
					Actions:    []string{"write", "read"},
					Identifier: clientAddress.String(),
				}},

			Label: "TestAccess1",
		}},
		Client: clientAddress.String(),
	})
	require.NoError(t, err)

	assert.Len(t, g.Tokens, 1)
	assert.Len(t, g.Tokens[0].Access, 2)
	for _, access := range g.Tokens[0].Access {
		if access.Type == "incoming-payment" {
			assert.EqualValues(t, access.Locations, []string{"https://ilp.link/incoming"})
			assert.EqualValues(t, access.Actions, []string{"write", "read"})
			continue
		}

		if access.Type == "outgoing-payment" {
			assert.EqualValues(t, access.Locations, []string{"https://ilp.link/outgoing"})
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
			assert.EqualValues(t, access.Locations, []string{"https://ilp.link/incoming"})
			assert.EqualValues(t, access.Actions, []string{"write", "read"})
			continue
		}

		if access.Type == "outgoing-payment" {
			assert.EqualValues(t, access.Locations, []string{"https://ilp.link/outgoing"})
			assert.EqualValues(t, access.Actions, []string{"write", "read"})
			continue
		}

		t.Fatal("Only expected outgoing-payment and incoming-payment access types")
	}
}
