package ops_test

import (
	"context"
	"testing"

	"gitlab.com/fynbos/backend/user"
	users_mock "gitlab.com/fynbos/backend/user/client/mock"

	"gitlab.com/fynbos/backend/limits"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gitlab.com/fynbos/backend/authorisation"
	auth_ops "gitlab.com/fynbos/backend/authorisation/ops"
	"gitlab.com/fynbos/backend/currency"
	"gitlab.com/fynbos/backend/db"
	"gitlab.com/fynbos/backend/limits/ops"
	"gitlab.com/fynbos/backend/transactions"
	tx_client "gitlab.com/fynbos/backend/transactions/client"
)

func TestExceeds(t *testing.T) {
	ctx := context.Background()
	dbc := db.MigrateTestDB(t, ctx)

	b := ops.NewTestBackends(t, dbc, users_mock.NewMock())
	txClient := tx_client.New(b)

	cases := []struct {
		name   string
		tx     transactions.CreateTransactionArgs
		amnt   currency.Amount
		expect bool
	}{
		{
			name: "exceeds daily defaults",
			tx: transactions.CreateTransactionArgs{
				State:       transactions.StateCompleted,
				Source:      "https://fynbos.me/alice1",
				Destination: "https://fynbos.me/bob1",
				Amount:      currency.FromFloat64(200, currency.USD),
				Provider:    "test",
				ForeignType: transactions.TransactionTypeOpenOutgoingPayment,
			},
			amnt:   currency.FromFloat64(200, currency.USD),
			expect: true,
		},
		{
			name: "does not exceed daily defaults",
			tx: transactions.CreateTransactionArgs{
				State:       transactions.StateCompleted,
				Source:      "https://fynbos.me/alice2",
				Destination: "https://fynbos.me/bob2",
				Amount:      currency.FromFloat64(2, currency.USD),
				Provider:    "test",
				ForeignType: transactions.TransactionTypeOpenOutgoingPayment,
			},
			amnt:   currency.FromFloat64(2, currency.USD),
			expect: false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			wallet, err := b.Users().CreateNewWallet(ctx, user.CreateWalletArgs{
				UserID: uuid.NewString(),
				Name:   "test",
			})
			require.NoError(t, err)

			client, err := auth_ops.CreateClient(ctx, b, tc.tx.Source)
			require.NoError(t, err)

			gid := uuid.NewString()
			_, err = b.DB().ExecContext(ctx, "INSERT INTO authorisation_grants (id, client_id, state, continue_token, wait) VALUES ($1, $2, $3, $4, $5)",
				gid, client.ID, authorisation.GrantStateApproved, uuid.NewString(), 0)
			require.NoError(t, err)

			tc.tx.WalletID = wallet.ID
			tc.tx.GrantID = gid

			_, err = txClient.CreateTransaction(ctx, tc.tx)
			require.NoError(t, err)

			exceeds, err := ops.Exceeds(ctx, b, wallet.ID, client.ID, tc.amnt)
			require.NoError(t, err)
			assert.Equal(t, tc.expect, exceeds)
		})
	}
}

func TestUpdateClientLimits(t *testing.T) {
	ctx := context.Background()
	dbc := db.MigrateTestDB(t, ctx)

	b := ops.NewTestBackends(t, dbc, users_mock.NewMock())

	wallet, err := b.Users().CreateNewWallet(ctx, user.CreateWalletArgs{
		UserID: uuid.NewString(),
		Name:   "test",
	})
	require.NoError(t, err)

	_, err = auth_ops.CreateClient(ctx, b, "https://fynbos.me/bobby")
	require.NoError(t, err)

	cases := []struct {
		name string
		wl   limits.Limit
	}{
		{
			name: "success new",
			wl: limits.Limit{
				Daily:   currency.FromFloat64(5, currency.USD),
				Monthly: currency.FromFloat64(20, currency.USD),
				Overall: currency.FromFloat64(200, currency.USD),
			},
		},
		{
			name: "success updated",
			wl: limits.Limit{
				Daily:   currency.FromFloat64(5, currency.USD),
				Monthly: currency.FromFloat64(3000, currency.USD),
				Overall: currency.FromFloat64(40000, currency.USD),
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := ops.UpdateClientLimits(ctx, b, wallet.ID, "https://fynbos.me/bobby", tc.wl)
			require.NoError(t, err)
		})
	}
}

func TestListLimits(t *testing.T) {
	ctx := context.Background()
	dbc := db.MigrateTestDB(t, ctx)

	b := ops.NewTestBackends(t, dbc, users_mock.NewMock())

	wallet, err := b.Users().CreateNewWallet(ctx, user.CreateWalletArgs{
		UserID: uuid.NewString(),
		Name:   "test",
	})
	require.NoError(t, err)

	_, err = auth_ops.CreateClient(ctx, b, "https://fynbos.me/bobby")
	require.NoError(t, err)

	cases := []struct {
		name string
		wl   limits.Limit
	}{
		{
			name: "success",
			wl: limits.Limit{
				Daily:   currency.FromFloat64(5, currency.USD),
				Monthly: currency.FromFloat64(20, currency.USD),
				Overall: currency.FromFloat64(200, currency.USD),
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err = ops.UpdateClientLimits(ctx, b, wallet.ID, "https://fynbos.me/bobby", tc.wl)
			require.NoError(t, err)

			l, err := ops.ListLimits(ctx, b, wallet.ID)
			require.NoError(t, err)

			assert.Len(t, l, 1)
			assert.Equal(t, "https://fynbos.me/bobby", l[0].ForeignDisplay)
			assert.Equal(t, limits.FKTypeClient, l[0].ForeignType)
			assert.Equal(t, uint64(5_00), l[0].Limit.Daily.Value)
			assert.Equal(t, uint64(20_00), l[0].Limit.Monthly.Value)
			assert.Equal(t, uint64(200_00), l[0].Limit.Overall.Value)
		})
	}
}

func TestExceedsGMTLimits(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	dbc := db.MigrateTestDB(t, ctx)

	b := ops.NewTestBackends(t, dbc, users_mock.NewMock())
	txClient := tx_client.New(b)

	cases := []struct {
		name   string
		tx     *transactions.CreateTransactionArgs
		amnt   currency.Amount
		expect bool
	}{
		{
			name:   "exceeds single tx limit",
			amnt:   currency.FromFloat64(11_000, currency.USD),
			expect: true,
		},
		{
			name: "does not exceed daily defaults",
			tx: &transactions.CreateTransactionArgs{
				State:       transactions.StateCompleted,
				Source:      "https://fynbos.me/alice2",
				Destination: "https://fynbos.me/bob2",
				Amount:      currency.FromFloat64(9_000, currency.USD),
				Provider:    "test",
				ForeignType: transactions.TransactionTypeOpenOutgoingPayment,
			},
			amnt:   currency.FromFloat64(200, currency.USD),
			expect: false,
		},
		{
			name: "does not exceed daily limits",
			tx: &transactions.CreateTransactionArgs{
				State:       transactions.StateCompleted,
				Source:      "https://fynbos.me/alice2",
				Destination: "https://fynbos.me/bob2",
				Amount:      currency.FromFloat64(9_000, currency.USD),
				Provider:    "test",
				ForeignType: transactions.TransactionTypeOpenOutgoingPayment,
			},
			amnt:   currency.FromFloat64(200, currency.USD),
			expect: false,
		},
		{
			name: "does exceed daily limits",
			tx: &transactions.CreateTransactionArgs{
				State:       transactions.StateCompleted,
				Source:      "https://fynbos.me/alice2",
				Destination: "https://fynbos.me/bob2",
				Amount:      currency.FromFloat64(9_000, currency.USD),
				Provider:    "test",
				ForeignType: transactions.TransactionTypeOpenOutgoingPayment,
			},
			amnt:   currency.FromFloat64(1_100, currency.USD),
			expect: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			wallet, err := b.Users().CreateNewWallet(ctx, user.CreateWalletArgs{
				UserID: uuid.NewString(),
				Name:   "test",
			})
			require.NoError(t, err)

			if tc.tx != nil {
				tc.tx.WalletID = wallet.ID

				_, err = txClient.CreateTransaction(ctx, *tc.tx)
				require.NoError(t, err)
			}

			exceeds, err := ops.ExceedsGMTLimits(ctx, b, wallet.ID, tc.amnt)
			require.NoError(t, err)
			assert.Equal(t, tc.expect, exceeds)
		})
	}
}
