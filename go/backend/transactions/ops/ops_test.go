package ops_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"gitlab.com/fynbos/backend/transactions"
	"gitlab.com/fynbos/backend/transactions/ops"
	users_client "gitlab.com/fynbos/backend/user/client"
	test_utils "gitlab.com/fynbos/backend/utils"
)

func TestCreateTransaction(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	dbc := test_utils.MigrateCockroachDB(t, ctx)

	b := ops.NewTestBackends(t, dbc)
	userClient := users_client.New(b, "fakeURL", "fakeAdminURL")

	cases := []struct {
		name string
		args transactions.CreateTransactionArgs
		err  error
	}{
		{
			name: "success",
			args: transactions.CreateTransactionArgs{
				WalletID:    uuid.NewString(),
				ForeignID:   uuid.NewString(),
				ForeignType: transactions.TransactionTypeOpenOutgoingPayment,
				Provider:    transactions.ProviderMachnet,
				State:       transactions.StatePending,
				Source:      "$fynbos.me/alice",
				Destination: "$fynbos.me/bob",
				Amount: transactions.Amount{
					Value:      1000,
					Asset:      "USD",
					AssetScale: 2,
				},
			},
		},
		{
			name: "success with transfers",
			args: transactions.CreateTransactionArgs{
				WalletID:    uuid.NewString(),
				ForeignID:   uuid.NewString(),
				ForeignType: transactions.TransactionTypeOpenOutgoingPayment,
				Provider:    transactions.ProviderMachnet,
				State:       transactions.StatePending,
				Source:      "$fynbos.me/alice",
				Destination: "$fynbos.me/bob",
				Amount: transactions.Amount{
					Value:      1000,
					Asset:      "USD",
					AssetScale: 2,
				},
				Transfers: []transactions.TransferArgs{
					{
						ForeignID: uuid.NewString(),
						Type:      transactions.TransferTypeDebitCard,
						State:     transactions.StateCompleted,
						Amount: transactions.Amount{
							Value:      1000,
							Asset:      "USD",
							AssetScale: 2,
						},
					},
					{
						ForeignID: uuid.NewString(),
						Type:      transactions.TransferTypeCreditMachnetWallet,
						State:     transactions.StateFailed,
						Amount: transactions.Amount{
							Value:      1000,
							Asset:      "USD",
							AssetScale: 2,
						},
					},
					{
						ForeignID: uuid.NewString(),
						Type:      transactions.TransferTypeDebitMachnetWallet,
						State:     transactions.StateFailed,
						Amount: transactions.Amount{
							Value:      1000,
							Asset:      "USD",
							AssetScale: 2,
						},
					},
				},
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			// Create Signups
			userID := uuid.NewString()
			_, err := dbc.ExecContext(ctx, "INSERT INTO signups (id, user_id) VALUES ($1, $2)", tc.args.WalletID, userID)
			require.NoError(t, err)
			// Create Wallets
			wallet, err := userClient.CreateNewWallet(ctx, userID, "test")
			require.NoError(t, err)

			tc.args.WalletID = wallet.ID

			err = ops.CreateTransaction(ctx, b, tc.args)
			require.NoError(t, err)
		})
	}
}

func TestUpdateTransfers(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	dbc := test_utils.MigrateCockroachDB(t, ctx)

	b := ops.NewTestBackends(t, dbc)
	userClient := users_client.New(b, "fakeURL", "fakeAdminURL")

	cases := []struct {
		name   string
		args   transactions.CreateTransactionArgs
		update transactions.TransferArgs
		err    error
	}{
		{
			name: "success",
			args: transactions.CreateTransactionArgs{
				WalletID:    uuid.NewString(),
				ForeignID:   "8ba075f3-f819-48bb-8e47-4f973acd4e72",
				ForeignType: transactions.TransactionTypeOpenOutgoingPayment,
				Provider:    transactions.ProviderMachnet,
				State:       transactions.StatePending,
				Source:      "$fynbos.me/alice",
				Destination: "$fynbos.me/bob",
				Amount: transactions.Amount{
					Value:      1000,
					Asset:      "USD",
					AssetScale: 2,
				},
				Transfers: []transactions.TransferArgs{
					{
						ForeignID: "0e4b5b26-a712-42e0-b7ea-f7870ca1b363",
						Type:      transactions.TransferTypeDebitCard,
						Amount: transactions.Amount{
							Value:      1000,
							Asset:      "USD",
							AssetScale: 2,
						},
						State: transactions.StatePending,
					},
				},
			},
			update: transactions.TransferArgs{
				WalletID:             uuid.NewString(),
				TransactionForeignID: "8ba075f3-f819-48bb-8e47-4f973acd4e72",
				ForeignID:            "0e4b5b26-a712-42e0-b7ea-f7870ca1b363",
				Type:                 transactions.TransferTypeDebitCard,
				Amount: transactions.Amount{
					Value:      1000,
					Asset:      "USD",
					AssetScale: 2,
				},
				State: transactions.StateCompleted,
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			// Create Signups
			userID := uuid.NewString()
			_, err := dbc.ExecContext(ctx, "INSERT INTO signups (id, user_id) VALUES ($1, $2)", tc.args.WalletID, userID)
			require.NoError(t, err)
			// Create Wallets
			wallet, err := userClient.CreateNewWallet(ctx, userID, "test")
			require.NoError(t, err)

			tc.args.WalletID = wallet.ID

			err = ops.CreateTransaction(ctx, b, tc.args)
			require.NoError(t, err)

			tc.update.WalletID = wallet.ID
			err = ops.UpdateTransfers(ctx, b, []transactions.TransferArgs{tc.update})
			require.NoError(t, err)
		})
	}
}
