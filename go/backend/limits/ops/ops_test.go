package ops_test

import (
	"context"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gitlab.com/fynbos/backend/authorisation"
	auth_ops "gitlab.com/fynbos/backend/authorisation/ops"
	"gitlab.com/fynbos/backend/currency"
	"gitlab.com/fynbos/backend/db"
	"gitlab.com/fynbos/backend/kyc"
	"gitlab.com/fynbos/backend/limits"
	"gitlab.com/fynbos/backend/limits/ops"
	"gitlab.com/fynbos/backend/transactions"
	tx_client "gitlab.com/fynbos/backend/transactions/client"
	users_mock "gitlab.com/fynbos/backend/user/client/mock"
)

func TestExceeds(t *testing.T) {
	ctx := context.Background()
	dbc := db.MigrateTestDB(t, ctx)

	b := NewTestBackends(t, dbc, users_mock.NewMock())
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
				Source:      "https://ilp.link/alice1",
				Destination: "https://ilp.link/bob1",
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
				Source:      "https://ilp.link/alice2",
				Destination: "https://ilp.link/bob2",
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
			walletID := uuid.NewString()

			client, err := auth_ops.CreateClient(ctx, b, tc.tx.Source)
			require.NoError(t, err)

			gid := uuid.NewString()
			_, err = b.DB().ExecContext(ctx, "INSERT INTO authorisation_grants (id, client_id, state, continue_token, wait) VALUES ($1, $2, $3, $4, $5)",
				gid, client.ID, authorisation.GrantStateApproved, uuid.NewString(), 0)
			require.NoError(t, err)

			tc.tx.WalletID = walletID
			tc.tx.GrantID = gid

			_, err = txClient.CreateTransaction(ctx, tc.tx)
			require.NoError(t, err)

			exceeds, err := ops.Exceeds(ctx, b, walletID, client.ID, tc.amnt)
			require.NoError(t, err)
			assert.Equal(t, tc.expect, exceeds)
		})
	}
}

func TestUpdateClientLimits(t *testing.T) {
	ctx := context.Background()
	dbc := db.MigrateTestDB(t, ctx)

	b := NewTestBackends(t, dbc, users_mock.NewMock())

	walletID := uuid.NewString()

	_, err := auth_ops.CreateClient(ctx, b, "https://ilp.link/bobby")
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
			err := ops.UpdateClientLimits(ctx, b, walletID, "https://ilp.link/bobby", tc.wl)
			require.NoError(t, err)
		})
	}
}

func TestListLimits(t *testing.T) {
	ctx := context.Background()
	dbc := db.MigrateTestDB(t, ctx)

	b := NewTestBackends(t, dbc, users_mock.NewMock())

	walletID := uuid.NewString()

	_, err := auth_ops.CreateClient(ctx, b, "https://ilp.link/bobby")
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
			err = ops.UpdateClientLimits(ctx, b, walletID, "https://ilp.link/bobby", tc.wl)
			require.NoError(t, err)

			l, err := ops.ListLimits(ctx, b, walletID)
			require.NoError(t, err)

			assert.Len(t, l, 1)
			assert.Equal(t, "https://ilp.link/bobby", l[0].ForeignDisplay)
			assert.Equal(t, limits.FKTypeClient, l[0].ForeignType)
			assert.Equal(t, uint64(5_00), l[0].Limit.Daily.Value)
			assert.Equal(t, uint64(20_00), l[0].Limit.Monthly.Value)
			assert.Equal(t, uint64(200_00), l[0].Limit.Overall.Value)
		})
	}
}

func TestExceedsKYCLimits(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	dbc := db.MigrateTestDB(t, ctx)

	b := NewTestBackends(t, dbc, users_mock.NewMock())
	txClient := tx_client.New(b)
	cases := []struct {
		name               string
		tx                 *transactions.CreateTransactionArgs
		amnt               currency.Amount
		expectExceedsLimit bool
		expectLimitType    limits.LimitType
		kycLevel           kyc.Status
	}{
		{
			name:               "exceeds single tx limit",
			amnt:               currency.FromFloat64(11_000, currency.USD),
			expectExceedsLimit: true,
			expectLimitType:    limits.LimitTypeTransaction,
			kycLevel:           kyc.StatusLevel1,
		},
		{
			name:               "under single tx limit no previous transactions",
			amnt:               currency.FromFloat64(249, currency.USD),
			expectExceedsLimit: false,
			kycLevel:           kyc.StatusLevel1,
		},
		{
			name: "does not exceed daily defaults",
			tx: &transactions.CreateTransactionArgs{
				State:       transactions.StateCompleted,
				Source:      "https://ilp.link/alice2",
				Destination: "https://ilp.link/bob2",
				Amount:      currency.FromFloat64(2_900, currency.USD),
				Provider:    "test",
				ForeignType: transactions.TransactionTypeOpenOutgoingPayment,
			},
			amnt:               currency.FromFloat64(98, currency.USD),
			expectExceedsLimit: false,
			kycLevel:           kyc.StatusLevel1,
		},
		{
			name: "does exceed daily limits",
			tx: &transactions.CreateTransactionArgs{
				State:       transactions.StateCompleted,
				Source:      "https://ilp.link/alice2",
				Destination: "https://ilp.link/bob2",
				Amount:      currency.FromFloat64(2_900, currency.USD),
				Provider:    "test",
				ForeignType: transactions.TransactionTypeOpenOutgoingPayment,
			},
			amnt:               currency.FromFloat64(99, currency.USD),
			expectExceedsLimit: true,
			expectLimitType:    limits.LimitTypeDaily,
			kycLevel:           kyc.StatusLevel1,
		},
		{
			name:               "pending kyc exceeds limits",
			amnt:               currency.FromFloat64(1, currency.USD),
			expectExceedsLimit: true,
			expectLimitType:    limits.LimitTypeTransaction,
			kycLevel:           kyc.StatusPending,
		},
		{
			name:               "denied kyc exceeds limits",
			amnt:               currency.FromFloat64(1, currency.USD),
			expectExceedsLimit: true,
			expectLimitType:    limits.LimitTypeTransaction,
			kycLevel:           kyc.StatusDenied,
		},
		{
			name:               "unknown kyc exceeds limits",
			amnt:               currency.FromFloat64(1, currency.USD),
			expectExceedsLimit: true,
			expectLimitType:    limits.LimitTypeTransaction,
			kycLevel:           kyc.StatusUnknown,
		},
		{
			name: "exceeds ZAR yearly limits",
			tx: &transactions.CreateTransactionArgs{
				State:       transactions.StateCompleted,
				Source:      "https://ilp.link/alice2",
				Destination: "https://ilp.link/bob2",
				Amount:      currency.FromFloat64(20_000, currency.ZAR),
				Provider:    "test",
				ForeignType: transactions.TransactionTypeOpenOutgoingPayment,
			},
			amnt:               currency.FromFloat64(99, currency.ZAR),
			expectExceedsLimit: true,
			expectLimitType:    limits.LimitTypeYearly,
			kycLevel:           kyc.StatusLevel1,
		},
		{
			name: "No ZAR limits",
			tx: &transactions.CreateTransactionArgs{
				State:       transactions.StateCompleted,
				Source:      "https://ilp.link/alice2",
				Destination: "https://ilp.link/bob2",
				Amount:      currency.FromFloat64(200_000, currency.ZAR),
				Provider:    "test",
				ForeignType: transactions.TransactionTypeOpenOutgoingPayment,
			},
			amnt:               currency.FromFloat64(99, currency.ZAR),
			expectExceedsLimit: false,
			kycLevel:           kyc.StatusLevel2,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			walletID := uuid.NewString()
			b.kyc.EXPECT().GetKYCStatus(gomock.Any(), walletID).Return(tc.kycLevel, nil).AnyTimes()

			if tc.tx != nil {
				tc.tx.WalletID = walletID

				_, err := txClient.CreateTransaction(ctx, *tc.tx)
				require.NoError(t, err)
			}

			exceeds, limitType, err := ops.ExceedsKYCLimits(ctx, b, walletID, tc.amnt)
			require.NoError(t, err)
			assert.Equal(t, tc.expectExceedsLimit, exceeds)
			assert.Equal(t, tc.expectLimitType, limitType)
		})
	}
}
