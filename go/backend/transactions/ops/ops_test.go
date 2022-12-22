package ops_test

import (
	"context"
	"testing"

	linkedaccounts_client "gitlab.com/fynbos/backend/linkedaccounts/client"

	"github.com/google/uuid"
	"gitlab.com/fynbos/backend/currency"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gitlab.com/fynbos/backend/db"
	"gitlab.com/fynbos/backend/linkedaccounts"
	"gitlab.com/fynbos/backend/transactions"
	"gitlab.com/fynbos/backend/transactions/ops"
	users_client "gitlab.com/fynbos/backend/user/client"
)

func TestCreateTransaction(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	dbc := db.MigrateTestDB(t, ctx)

	b := ops.NewTestBackends(t, dbc)
	userClient := users_client.New(b, "fakeURL", "fakeAdminURL")
	laClient := linkedaccounts_client.New(b)

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
				Amount: currency.Amount{
					Value:    1000,
					Currency: currency.USD,
					Scale:    2,
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
				Amount: currency.Amount{
					Value:    1000,
					Currency: currency.USD,
					Scale:    2,
				},
				Transfers: []transactions.TransferArgs{
					{
						ForeignID: uuid.NewString(),
						Type:      transactions.TransferTypeDebitCard,
						State:     transactions.StateCompleted,
						Amount: currency.Amount{
							Value:    1000,
							Currency: currency.USD,
							Scale:    2,
						},
					},
					{
						ForeignID: uuid.NewString(),
						Type:      transactions.TransferTypeCreditMachnetWallet,
						State:     transactions.StateFailed,
						Amount: currency.Amount{
							Value:    1000,
							Currency: currency.USD,
							Scale:    2,
						},
					},
					{
						ForeignID: uuid.NewString(),
						Type:      transactions.TransferTypeDebitMachnetWallet,
						State:     transactions.StateFailed,
						Amount: currency.Amount{
							Value:    1000,
							Currency: currency.USD,
							Scale:    2,
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

			la, err := laClient.Create(ctx, &linkedaccounts.CreateArgs{
				WalletID:   wallet.ID,
				Name:       "test",
				Mask:       "ladida",
				Provider:   "machnet",
				ProviderID: uuid.NewString(),
				Type:       "test",
			})
			require.NoError(t, err)

			tc.args.WalletID = wallet.ID
			for i := range tc.args.Transfers {
				tc.args.Transfers[i].LinkedAccountID = la.ID
			}

			err = ops.CreateTransaction(ctx, b, tc.args)
			require.NoError(t, err)
		})
	}
}

func TestUpdateTransfers(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	dbc := db.MigrateTestDB(t, ctx)

	b := ops.NewTestBackends(t, dbc)
	userClient := users_client.New(b, "fakeURL", "fakeAdminURL")
	laClient := linkedaccounts_client.New(b)

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
				Amount: currency.Amount{
					Value:    1000,
					Currency: currency.USD,
					Scale:    2,
				},
				Transfers: []transactions.TransferArgs{
					{
						ForeignID: "0e4b5b26-a712-42e0-b7ea-f7870ca1b363",
						Type:      transactions.TransferTypeDebitCard,
						Amount: currency.Amount{
							Value:    1000,
							Currency: currency.USD,
							Scale:    2,
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
				Amount: currency.Amount{
					Value:    1000,
					Currency: currency.USD,
					Scale:    2,
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
			la, err := laClient.Create(ctx, &linkedaccounts.CreateArgs{
				WalletID:   wallet.ID,
				Name:       "test",
				Mask:       "ladida",
				Provider:   "machnet",
				ProviderID: uuid.NewString(),
				Type:       "test",
			})
			require.NoError(t, err)

			tc.args.WalletID = wallet.ID
			for i := range tc.args.Transfers {
				tc.args.Transfers[i].LinkedAccountID = la.ID
			}

			err = ops.CreateTransaction(ctx, b, tc.args)
			require.NoError(t, err)

			tc.update.WalletID = wallet.ID
			err = ops.UpdateTransfers(ctx, b, []transactions.TransferArgs{tc.update})
			require.NoError(t, err)
		})
	}
}

func TestListTransaction(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	dbc := db.MigrateTestDB(t, ctx)

	b := ops.NewTestBackends(t, dbc)
	userClient := users_client.New(b, "fakeURL", "fakeAdminURL")
	laClient := linkedaccounts_client.New(b)

	cases := []struct {
		name string
		args transactions.CreateTransactionArgs
		len  int
	}{
		{
			name: "no transfers",
			len:  1,
			args: transactions.CreateTransactionArgs{
				WalletID:    uuid.NewString(),
				ForeignID:   uuid.NewString(),
				ForeignType: transactions.TransactionTypeOpenOutgoingPayment,
				Provider:    transactions.ProviderMachnet,
				State:       transactions.StatePending,
				Source:      "$fynbos.me/alice",
				Destination: "$fynbos.me/bob",
				Amount: currency.Amount{
					Value:    1000,
					Currency: currency.USD,
					Scale:    2,
				},
			},
		},
		{
			name: "ignore incoming openpayments pending",
			len:  0,
			args: transactions.CreateTransactionArgs{
				WalletID:    uuid.NewString(),
				ForeignID:   uuid.NewString(),
				ForeignType: transactions.TransactionTypeOpenPaymentIncoming,
				Provider:    transactions.ProviderMachnet,
				State:       transactions.StatePending,
				Source:      "$fynbos.me/alice",
				Destination: "$fynbos.me/bob",
				Amount: currency.Amount{
					Value:    1000,
					Currency: currency.USD,
					Scale:    2,
				},
			},
		},
		{
			name: "with transfers",
			len:  1,
			args: transactions.CreateTransactionArgs{
				WalletID:    uuid.NewString(),
				ForeignID:   uuid.NewString(),
				ForeignType: transactions.TransactionTypeOpenOutgoingPayment,
				Provider:    transactions.ProviderMachnet,
				State:       transactions.StatePending,
				Source:      "$fynbos.me/alice",
				Destination: "$fynbos.me/bob",
				Amount: currency.Amount{
					Value:    1000,
					Currency: currency.USD,
					Scale:    2,
				},
				Transfers: []transactions.TransferArgs{
					{
						ForeignID: uuid.NewString(),
						Type:      transactions.TransferTypeDebitCard,
						State:     transactions.StateCompleted,
						Amount: currency.Amount{
							Value:    1000,
							Currency: currency.USD,
							Scale:    2,
						},
					},
					{
						ForeignID: uuid.NewString(),
						Type:      transactions.TransferTypeCreditMachnetWallet,
						State:     transactions.StateFailed,
						Amount: currency.Amount{
							Value:    1000,
							Currency: currency.USD,
							Scale:    2,
						},
					},
					{
						ForeignID: uuid.NewString(),
						Type:      transactions.TransferTypeDebitMachnetWallet,
						State:     transactions.StateFailed,
						Amount: currency.Amount{
							Value:    1000,
							Currency: currency.USD,
							Scale:    2,
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
			la, err := laClient.Create(ctx, &linkedaccounts.CreateArgs{
				WalletID:   wallet.ID,
				Name:       "test",
				Mask:       "ladida",
				Provider:   "machnet",
				ProviderID: uuid.NewString(),
				Type:       "test",
			})
			require.NoError(t, err)

			tc.args.WalletID = wallet.ID
			for i := range tc.args.Transfers {
				tc.args.Transfers[i].LinkedAccountID = la.ID
			}

			err = ops.CreateTransaction(ctx, b, tc.args)
			require.NoError(t, err)

			txs, err := ops.ListTransactions(ctx, b, wallet.ID, db.Pagination{})
			require.NoError(t, err)
			require.Len(t, txs, tc.len)

			if len(txs) == 0 {
				return
			}

			for _, tx := range txs {
				assert.Equal(t, tc.args.State, tx.State)
				assert.Equal(t, tc.args.Provider, tx.Provider)
				assert.Equal(t, tc.args.ForeignType, tx.Type)
				assert.Equal(t, tc.args.Source, tx.Source)
				assert.Equal(t, tc.args.Destination, tx.Destination)
				assert.Equal(t, tc.args.Note, tx.Note)
				for _, tr := range tx.Transfers {
					var found bool
					for _, etr := range tc.args.Transfers {
						if etr.State == tr.State && etr.Type == tr.Type {
							found = true
							break
						}
					}
					assert.True(t, found)
				}
			}
		})
	}
}
