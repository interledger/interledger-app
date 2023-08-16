package ops_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gitlab.com/fynbos/backend/currency"
	"gitlab.com/fynbos/backend/db"
	"gitlab.com/fynbos/backend/payments"
	"gitlab.com/fynbos/backend/payments/ops"
)

func TestCreate(t *testing.T) {
	ctx := context.Background()

	b := ops.NewTestBackends(t, func(b *ops.TestBackends) {
		b.DBC = db.MigrateTestDB(t, ctx)
	})

	cases := []struct {
		name    string
		args    payments.CreateArgs
		actions []payments.RequiredActionType
		err     error
	}{
		{
			name: "success",
			args: payments.CreateArgs{
				Sender: payments.Identity{
					Type:       payments.IdentityTypeTwitter,
					Identifier: "@willy_wonka",
				},
				Receiver: payments.Identity{
					Type:       payments.IdentityTypeWalletURL,
					Identifier: "https://fynbos.me/charlie",
				},
				SenderAmount:    currency.FromFloat64(51, currency.USD),
				ReceiverAmount:  currency.FromFloat64(50, currency.USD),
				SenderAccount:   uuid.NewString(),
				ReceiverAccount: uuid.NewString(),
				Note:            "This is a NOTE!!!",
			},
			actions: []payments.RequiredActionType{payments.RequiredActionTypeThreeDS},
			err:     nil,
		},
		{
			name: "success_no_accounts",
			args: payments.CreateArgs{
				Sender: payments.Identity{
					Type:       payments.IdentityTypeTwitter,
					Identifier: "@willy_wonka",
				},
				Receiver: payments.Identity{
					Type:       payments.IdentityTypeWalletURL,
					Identifier: "https://fynbos.me/charlie",
				},
				SenderAmount:   currency.FromFloat64(51, currency.USD),
				ReceiverAmount: currency.FromFloat64(50, currency.USD),
			},
			actions: []payments.RequiredActionType{payments.RequiredActionTypeThreeDS},
			err:     nil,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			p, err := ops.Create(ctx, b, tc.args)
			if tc.err != nil {
				require.ErrorIs(t, err, tc.err)
				return
			}
			require.NoError(t, err)

			assert.Equal(t, tc.args.Sender.Type, p.Sender.Type)
			assert.Equal(t, tc.args.Sender.Identifier, p.Sender.Identifier)
			assert.Equal(t, tc.args.Receiver.Type, p.Receiver.Type)
			assert.Equal(t, tc.args.Receiver.Identifier, p.Receiver.Identifier)
			assert.Equal(t, tc.args.SenderAccount, p.SenderAccount)
			assert.Equal(t, tc.args.ReceiverAccount, p.ReceiverAccount)
			assert.Equal(t, tc.args.SenderAmount.Format(), p.SenderAmount.Format())
			assert.Equal(t, tc.args.ReceiverAmount.Format(), p.ReceiverAmount.Format())
			assert.Equal(t, tc.args.Note, p.Note)
			assert.Len(t, p.RequiredActions, len(tc.actions))
			for _, ra := range tc.actions {
				assert.Contains(t, p.RequiredActions, ra)
			}
		})
	}
}

func TestSetState(t *testing.T) {
	ctx := context.Background()

	b := ops.NewTestBackends(t, func(b *ops.TestBackends) {
		b.DBC = db.MigrateTestDB(t, ctx)
	})

	p, err := ops.Create(ctx, b, payments.CreateArgs{
		Sender: payments.Identity{
			Type:       payments.IdentityTypeTwitter,
			Identifier: "@willy_wonka",
		},
		Receiver: payments.Identity{
			Type:       payments.IdentityTypeWalletURL,
			Identifier: "https://fynbos.me/charlie",
		},
		SenderAmount:    currency.FromFloat64(51, currency.USD),
		ReceiverAmount:  currency.FromFloat64(50, currency.USD),
		SenderAccount:   uuid.NewString(),
		ReceiverAccount: uuid.NewString(),
	})
	require.NoError(t, err)

	assert.ErrorIs(t, ops.SetState(ctx, b, p.ID, payments.StateCompleted), payments.ErrInvalidStateTransition)
	assert.NoError(t, ops.SetState(ctx, b, p.ID, payments.StateConfirmed))
}
