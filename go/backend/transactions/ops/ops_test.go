package ops_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gitlab.com/fynbos/backend/currency"
	"gitlab.com/fynbos/backend/db"
	"gitlab.com/fynbos/backend/linkedaccounts"
	linkedaccounts_client "gitlab.com/fynbos/backend/linkedaccounts/client"
	"gitlab.com/fynbos/backend/transactions"
	"gitlab.com/fynbos/backend/transactions/ops"
	"gitlab.com/fynbos/backend/user"
	users_mock "gitlab.com/fynbos/backend/user/client/mock"
)

func TestCreateTransaction(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	dbc := db.MigrateTestDB(t, ctx)

	b := ops.NewTestBackends(t, dbc)
	userClient := users_mock.NewMock()
	laClient := linkedaccounts_client.New(b)

	cases := []struct {
		name string
		args transactions.CreateTransactionArgs
		err  error
	}{
		{
			name: "success with provided ID",
			args: transactions.CreateTransactionArgs{
				ID:          uuid.NewString(),
				WalletID:    uuid.NewString(),
				ForeignID:   uuid.NewString(),
				ForeignType: transactions.TransactionTypeOpenOutgoingPayment,
				Provider:    transactions.ProviderGMT,
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
			name: "success with foreign id",
			args: transactions.CreateTransactionArgs{
				WalletID:    uuid.NewString(),
				ForeignID:   uuid.NewString(),
				ForeignType: transactions.TransactionTypeOpenOutgoingPayment,
				Provider:    transactions.ProviderGMT,
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
			name: "success without foreignID",
			args: transactions.CreateTransactionArgs{
				WalletID:    uuid.NewString(),
				ForeignType: transactions.TransactionTypeOpenOutgoingPayment,
				Provider:    transactions.ProviderGMT,
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
				Provider:    transactions.ProviderGMT,
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
						Type:      transactions.TransferTypeCreditBankAccount,
						State:     transactions.StateFailed,
						Amount: currency.Amount{
							Value:    1000,
							Currency: currency.USD,
							Scale:    2,
						},
					},
					{
						ForeignID: uuid.NewString(),
						Type:      transactions.TransferTypeDebitBankAccount,
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
			// Create Wallets
			userID := uuid.NewString()
			wallet, err := userClient.CreateNewWallet(ctx, user.CreateWalletArgs{
				UserID: userID,
				Name:   "test",
			})
			require.NoError(t, err)

			la, err := laClient.Create(ctx, &linkedaccounts.CreateArgs{
				WalletID:   wallet.ID,
				Name:       "test",
				Mask:       "ladida",
				Provider:   "gmt",
				ProviderID: uuid.NewString(),
				Type:       "test",
			})
			require.NoError(t, err)

			tc.args.WalletID = wallet.ID
			for i := range tc.args.Transfers {
				tc.args.Transfers[i].LinkedAccountID = la.ID
			}

			tid, err := ops.CreateTransaction(ctx, b, tc.args)
			require.NoError(t, err)
			if tc.args.ID != "" {
				assert.Equal(t, tc.args.ID, tid)
			}
		})
	}
}

func TestListWithPendingTransaction(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	dbc := db.MigrateTestDB(t, ctx)

	b := ops.NewTestBackends(t, dbc)
	userClient := users_mock.NewMock()
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
				Provider:    transactions.ProviderGMT,
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
				Provider:    transactions.ProviderGMT,
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
				Provider:    transactions.ProviderGMT,
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
						Type:      transactions.TransferTypeCreditBankAccount,
						State:     transactions.StateFailed,
						Amount: currency.Amount{
							Value:    1000,
							Currency: currency.USD,
							Scale:    2,
						},
					},
					{
						ForeignID: uuid.NewString(),
						Type:      transactions.TransferTypeDebitBankAccount,
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
			// Create Wallets
			userID := uuid.NewString()
			wallet, err := userClient.CreateNewWallet(ctx, user.CreateWalletArgs{
				UserID: userID,
				Name:   "test",
			})
			require.NoError(t, err)

			tc.args.WalletID = wallet.ID
			la, err := laClient.Create(ctx, &linkedaccounts.CreateArgs{
				WalletID:   wallet.ID,
				Name:       "test",
				Mask:       "ladida",
				Provider:   "gmt",
				ProviderID: uuid.NewString(),
				Type:       "test",
			})
			require.NoError(t, err)

			tc.args.WalletID = wallet.ID
			for i := range tc.args.Transfers {
				tc.args.Transfers[i].LinkedAccountID = la.ID
			}

			_, err = ops.CreateTransaction(ctx, b, tc.args)
			require.NoError(t, err)

			txs, err := ops.ListWithPending(ctx, b, wallet.ID, db.Pagination{})
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

func TestListWithPendingPagination(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	dbc := db.MigrateTestDB(t, ctx)

	b := ops.NewTestBackends(t, dbc)
	userClient := users_mock.NewMock()

	pendingTxs := make([]transactions.CreateTransactionArgs, 20)
	for i := range pendingTxs {
		pendingTxs[i] = transactions.CreateTransactionArgs{
			WalletID:    uuid.NewString(),
			ForeignID:   uuid.NewString(),
			ForeignType: transactions.TransactionTypeOpenOutgoingPayment,
			Provider:    transactions.ProviderGMT,
			State:       transactions.StatePending,
			Source:      "$fynbos.me/alice",
			Destination: "$fynbos.me/bob",
			Amount: currency.Amount{
				Value:    1000,
				Currency: currency.USD,
				Scale:    2,
			},
		}
	}

	cases := []struct {
		name  string
		args  db.Pagination
		start int
		len   int
	}{
		{
			name:  "no pagination",
			len:   20,
			start: 0,
			args:  db.Pagination{},
		},
		{
			name:  "No overflow returns PageSize+1 transactions",
			len:   5,
			start: 0,
			args: db.Pagination{
				PageSize:  4,
				PageToken: "",
			},
		},
		{
			name:  "Can paginate starting at a specific token",
			len:   10,
			start: 1,
			args: db.Pagination{
				PageSize:  9,
				PageToken: "",
			},
		},
		{
			name:  "Can paginate starting at a specific token with overflow",
			len:   15,
			start: 5,
			args: db.Pagination{
				PageSize:  20,
				PageToken: "",
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			// Create Signups
			userID := uuid.NewString()
			// Create Wallets
			wallet, err := userClient.CreateNewWallet(ctx, user.CreateWalletArgs{
				UserID: userID,
				Name:   "test",
			})
			require.NoError(t, err)

			for i, tx := range pendingTxs {
				tx.WalletID = wallet.ID
				txId, err := ops.CreateTransaction(ctx, b, tx)
				require.NoError(t, err)
				if tc.start != 0 && tc.start == len(pendingTxs)-i-1 {
					tc.args.PageToken = txId
				}
			}

			txs, err := ops.ListWithPending(ctx, b, wallet.ID, tc.args)
			require.NoError(t, err)
			require.Len(t, txs, tc.len)

			require.GreaterOrEqual(t, txs[0].Timestamp, txs[len(txs)-1].Timestamp)

			if tc.start != 0 {
				require.Equal(t, tc.args.PageToken, txs[0].ID)
			}
		})
	}
}

func TestSetTransactionForeignIDs(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	dbc := db.MigrateTestDB(t, ctx)

	b := ops.NewTestBackends(t, dbc)
	userClient := users_mock.NewMock()
	laClient := linkedaccounts_client.New(b)

	cases := []struct {
		name      string
		args      transactions.CreateTransactionArgs
		foreignID string
	}{
		{
			name: "success",
			args: transactions.CreateTransactionArgs{
				WalletID:    uuid.NewString(),
				ForeignType: transactions.TransactionTypeOpenOutgoingPayment,
				Provider:    transactions.ProviderGMT,
				State:       transactions.StatePending,
				Source:      "$fynbos.me/alice",
				Destination: "$fynbos.me/bob",
				Amount: currency.Amount{
					Value:    1000,
					Currency: currency.USD,
					Scale:    2,
				},
			},
			foreignID: "4e57c03d-90f2-4555-b5e4-5f07c4e40583",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			// Create Signups
			userID := uuid.NewString()
			// Create Wallets
			wallet, err := userClient.CreateNewWallet(ctx, user.CreateWalletArgs{
				UserID: userID,
				Name:   "test",
			})
			require.NoError(t, err)

			tc.args.WalletID = wallet.ID
			la, err := laClient.Create(ctx, &linkedaccounts.CreateArgs{
				WalletID:   wallet.ID,
				Name:       "test",
				Mask:       "ladida",
				Provider:   "gmt",
				ProviderID: uuid.NewString(),
				Type:       "test",
			})
			require.NoError(t, err)

			tc.args.WalletID = wallet.ID
			for i := range tc.args.Transfers {
				tc.args.Transfers[i].LinkedAccountID = la.ID
			}

			trxID, err := ops.CreateTransaction(ctx, b, tc.args)
			require.NoError(t, err)

			tx, err := ops.GetTransaction(ctx, b, wallet.ID, trxID)
			require.NoError(t, err)
			require.Empty(t, tx.ForeignID)

			err = ops.SetTransactionForeignID(ctx, b, trxID, tc.foreignID)
			require.NoError(t, err)

			tx, err = ops.GetTransaction(ctx, b, wallet.ID, trxID)
			require.NoError(t, err)

			assert.Equal(t, tx.ForeignID, tc.foreignID)
		})
	}
}

func TestSetTransferForeignID(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	dbc := db.MigrateTestDB(t, ctx)

	b := ops.NewTestBackends(t, dbc)
	userClient := users_mock.NewMock()
	laClient := linkedaccounts_client.New(b)

	cases := []struct {
		name      string
		args      transactions.CreateTransactionArgs
		foreignID string
	}{
		{
			name: "success",
			args: transactions.CreateTransactionArgs{
				WalletID:    uuid.NewString(),
				ForeignType: transactions.TransactionTypeOpenOutgoingPayment,
				Provider:    transactions.ProviderGMT,
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
						Type:  transactions.TransferTypeDebitCard,
						State: transactions.StatePending,
						Amount: currency.Amount{
							Value:    1000,
							Currency: currency.USD,
							Scale:    2,
						},
					},
				},
			},
			foreignID: "4e57c03d-90f2-4555-b5e4-5f07c4e40583",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			// Create Signups
			userID := uuid.NewString()
			// Create Wallets
			wallet, err := userClient.CreateNewWallet(ctx, user.CreateWalletArgs{
				UserID: userID,
				Name:   "test",
			})
			require.NoError(t, err)

			tc.args.WalletID = wallet.ID
			la, err := laClient.Create(ctx, &linkedaccounts.CreateArgs{
				WalletID:   wallet.ID,
				Name:       "test",
				Mask:       "ladida",
				Provider:   "gmt",
				ProviderID: uuid.NewString(),
				Type:       "test",
			})
			require.NoError(t, err)

			tc.args.WalletID = wallet.ID
			for i := range tc.args.Transfers {
				tc.args.Transfers[i].LinkedAccountID = la.ID
			}

			trxID, err := ops.CreateTransaction(ctx, b, tc.args)
			require.NoError(t, err)

			tx, err := ops.GetTransaction(ctx, b, wallet.ID, trxID)
			require.NoError(t, err)
			require.Empty(t, tx.ForeignID)

			tfr := tx.Transfers[0]
			require.Empty(t, tfr.ForeignID)

			fmt.Println(tfr.ID)
			err = ops.SetTransferForeignID(ctx, b, tfr.ID, tc.foreignID)
			require.NoError(t, err)

			tx, err = ops.GetTransaction(ctx, b, wallet.ID, trxID)
			require.NoError(t, err)

			assert.Equal(t, tx.Transfers[0].ForeignID, tc.foreignID)
		})
	}
}

func TestSetTransactionState(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	dbc := db.MigrateTestDB(t, ctx)

	b := ops.NewTestBackends(t, dbc)
	userClient := users_mock.NewMock()
	laClient := linkedaccounts_client.New(b)

	cases := []struct {
		name  string
		args  transactions.CreateTransactionArgs
		state transactions.State
	}{
		{
			name: "success",
			args: transactions.CreateTransactionArgs{
				WalletID:    uuid.NewString(),
				ForeignID:   uuid.NewString(),
				ForeignType: transactions.TransactionTypeOpenOutgoingPayment,
				Provider:    transactions.ProviderGMT,
				State:       transactions.StatePending,
				Source:      "$fynbos.me/alice",
				Destination: "$fynbos.me/bob",
				Amount: currency.Amount{
					Value:    1000,
					Currency: currency.USD,
					Scale:    2,
				},
			},
			state: transactions.StateCompleted,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			// Create Signups
			userID := uuid.NewString()
			// Create Wallets
			wallet, err := userClient.CreateNewWallet(ctx, user.CreateWalletArgs{
				UserID: userID,
				Name:   "test",
			})
			require.NoError(t, err)

			tc.args.WalletID = wallet.ID
			la, err := laClient.Create(ctx, &linkedaccounts.CreateArgs{
				WalletID:   wallet.ID,
				Name:       "test",
				Mask:       "ladida",
				Provider:   "gmt",
				ProviderID: uuid.NewString(),
				Type:       "test",
			})
			require.NoError(t, err)

			tc.args.WalletID = wallet.ID
			for i := range tc.args.Transfers {
				tc.args.Transfers[i].LinkedAccountID = la.ID
			}

			trxID, err := ops.CreateTransaction(ctx, b, tc.args)
			require.NoError(t, err)

			tx, err := ops.GetTransaction(ctx, b, wallet.ID, trxID)
			require.NoError(t, err)
			require.Equal(t, tx.State, tc.args.State)

			err = ops.SetTransactionState(ctx, b, trxID, tc.state)
			require.NoError(t, err)

			tx, err = ops.GetTransaction(ctx, b, wallet.ID, trxID)
			require.NoError(t, err)

			assert.Equal(t, tx.State, tc.state)
		})
	}
}

func TestSetTransferState(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	dbc := db.MigrateTestDB(t, ctx)

	b := ops.NewTestBackends(t, dbc)
	userClient := users_mock.NewMock()
	laClient := linkedaccounts_client.New(b)

	cases := []struct {
		name  string
		args  transactions.CreateTransactionArgs
		state transactions.State
	}{
		{
			name: "success",
			args: transactions.CreateTransactionArgs{
				WalletID:    uuid.NewString(),
				ForeignType: transactions.TransactionTypeOpenOutgoingPayment,
				Provider:    transactions.ProviderGMT,
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
						Type:  transactions.TransferTypeDebitCard,
						State: transactions.StatePending,
						Amount: currency.Amount{
							Value:    1000,
							Currency: currency.USD,
							Scale:    2,
						},
					},
				},
			},
			state: transactions.StateCompleted,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			// Create Signups
			userID := uuid.NewString()
			// Create Wallets
			wallet, err := userClient.CreateNewWallet(ctx, user.CreateWalletArgs{
				UserID: userID,
				Name:   "test",
			})
			require.NoError(t, err)

			tc.args.WalletID = wallet.ID
			la, err := laClient.Create(ctx, &linkedaccounts.CreateArgs{
				WalletID:   wallet.ID,
				Name:       "test",
				Mask:       "ladida",
				Provider:   "gmt",
				ProviderID: uuid.NewString(),
				Type:       "test",
			})
			require.NoError(t, err)

			tc.args.WalletID = wallet.ID
			for i := range tc.args.Transfers {
				tc.args.Transfers[i].LinkedAccountID = la.ID
			}

			trxID, err := ops.CreateTransaction(ctx, b, tc.args)
			require.NoError(t, err)

			tx, err := ops.GetTransaction(ctx, b, wallet.ID, trxID)
			require.NoError(t, err)
			tfr := tx.Transfers[0]
			require.Equal(t, tfr.State, tc.args.Transfers[0].State)

			err = ops.SetTransferState(ctx, b, tfr.ID, tc.state)
			require.NoError(t, err)

			tx, err = ops.GetTransaction(ctx, b, wallet.ID, trxID)
			require.NoError(t, err)

			assert.Equal(t, tx.Transfers[0].State, tc.state)
		})
	}
}
