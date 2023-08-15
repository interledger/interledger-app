package ops_test

import (
	"context"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gitlab.com/fynbos/backend/currency"
	"gitlab.com/fynbos/backend/db"
	"gitlab.com/fynbos/backend/payments"
	"gitlab.com/fynbos/backend/payments/ops"
	temporal_mock "gitlab.com/fynbos/backend/temporal/mock"
	temporal_client "go.temporal.io/sdk/client"
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

func TestGetRequiredActions(t *testing.T) {
	ctx := context.Background()
	b := ops.NewTestBackends(t, func(b *ops.TestBackends) {
		b.DBC = db.MigrateTestDB(t, ctx)
	})

	p, err := ops.Create(ctx, b, payments.CreateArgs{
		Sender: payments.Identity{
			Type:       payments.IdentityTypeTwitter,
			Identifier: "@willy_wonka",
		},
	})
	require.NoError(t, err)

	requiredActions, err := ops.GetRequiredActions(ctx, b, p.ID)
	require.NoError(t, err)
	assert.Contains(t, requiredActions, payments.RequiredActionTypeReceiverAmount)
	assert.Contains(t, requiredActions, payments.RequiredActionTypeReceiverIdentifier)
	assert.Contains(t, requiredActions, payments.RequiredActionTypeSenderAmount)
	assert.Contains(t, requiredActions, payments.RequiredActionTypeSenderAccount)
}

func TestConfirm(t *testing.T) {
	ctx := context.Background()
	ctrl := gomock.NewController(t)
	t.Cleanup(func() {
		ctrl.Finish()
	})
	b := &ops.TestBackends{
		DBC: db.MigrateTestDB(t, ctx),
		Tp:  temporal_mock.NewMockClient(ctrl),
	}

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
	b.Tp.EXPECT().ExecuteWorkflow(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(
		func(ctx context.Context, wo temporal_client.StartWorkflowOptions, w interface{}, args ...interface{}) (temporal_client.WorkflowRun, error) {
			assert.Len(t, args, 1)
			assert.Equal(t, p.ID, args[0])
			return nil, nil
		},
	).Times(1)

	p, requiredActions, err := ops.Confirm(ctx, b, p.ID)
	require.NoError(t, err)
	assert.Empty(t, requiredActions)
	assert.Equal(t, payments.StateConfirmed, p.State)
}
