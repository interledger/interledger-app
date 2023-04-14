package ops_test

import (
	"context"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gitlab.com/fynbos/backend/authorisation"
	"gitlab.com/fynbos/backend/authorisation/ops"
	"gitlab.com/fynbos/backend/db"
	"gitlab.com/fynbos/backend/openpayments"
	openpayments_mock "gitlab.com/fynbos/backend/openpayments/client/mock"
)

func TestCreateClient(t *testing.T) {
	cases := []struct {
		name          string
		pointer       string
		pointerExists bool
		err           error
	}{
		{
			name:          "success",
			pointer:       "https://fynbos.me/tables",
			pointerExists: true,
			err:           nil,
		},
		{
			name:          "does not exist",
			pointer:       "https://fynbos.me/notehere",
			pointerExists: false,
			err:           openpayments.ErrPaymentPointerNotFound,
		},
	}

	ctx := context.Background()

	ctrl := gomock.NewController(t)
	op := openpayments_mock.NewMockClient(ctrl)

	b := ops.NewTestBackends(t, db.MigrateTestDB(t, ctx), op)

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.pointerExists {
				op.EXPECT().GetPaymentPointer(ctx, tc.pointer).Return(&openpayments.PaymentPointer{URL: tc.pointer}, nil).Times(1)
			} else {
				op.EXPECT().GetPaymentPointer(ctx, tc.pointer).Return(nil, openpayments.ErrPaymentPointerNotFound).Times(1)
			}
			_, err := ops.CreateClient(ctx, b, tc.pointer)
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
	op := openpayments_mock.NewMockClient(ctrl)

	b := ops.NewTestBackends(t, db.MigrateTestDB(t, ctx), op)
	clientPaymentPointer := "https://fynbos.me/alice"
	resourceOwnerPaymentPointer := "https://fynbos.me/bobby"

	op.EXPECT().GetPaymentPointer(ctx, clientPaymentPointer).Return(&openpayments.PaymentPointer{URL: clientPaymentPointer}, nil).AnyTimes()
	op.EXPECT().GetPaymentPointer(ctx, resourceOwnerPaymentPointer).Return(&openpayments.PaymentPointer{URL: resourceOwnerPaymentPointer}, nil).AnyTimes()

	_, err := ops.CreateClient(ctx, b, clientPaymentPointer)
	require.NoError(t, err)

	g, err := ops.CreateGrant(ctx, b, authorisation.GrantRequest{
		AccessToken: []authorisation.AccessTokenReq{{
			Access: []authorisation.Access{
				{
					// To be ignored, you can only get tokens for yourself, currently
					Type:       "incoming-payment",
					Actions:    []string{"write", "read"},
					Identifier: resourceOwnerPaymentPointer,
				},
				{
					// To be ignored, you can only get tokens for yourself, currently
					Type:       "outgoing-payment",
					Actions:    []string{"write", "read"},
					Identifier: resourceOwnerPaymentPointer,
				},
				{
					Type:       "incoming-payment",
					Actions:    []string{"write", "read"},
					Identifier: clientPaymentPointer,
				},
				{
					Type:       "outgoing-payment",
					Actions:    []string{"write", "read"},
					Identifier: clientPaymentPointer,
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
