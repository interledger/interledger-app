package ops_test

import (
	"context"
	"testing"

	"gitlab.com/fynbos/backend/providers/pti"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gitlab.com/fynbos/backend/currency"
	"gitlab.com/fynbos/backend/db"
	"gitlab.com/fynbos/backend/linkedaccounts"
	linkedaccounts_client "gitlab.com/fynbos/backend/linkedaccounts/client"
	"gitlab.com/fynbos/backend/transactions"
	"gitlab.com/fynbos/backend/transactions/ops"
)

func TestCreateTransaction(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	dbc := db.MigrateTestDB(t, ctx)

	b := ops.NewTestBackends(t, dbc)
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
				Provider:    transactions.ProviderPaymentsEngine,
				State:       transactions.StatePending,
				Source:      "$ilp.link/alice",
				Destination: "$ilp.link/bob",
				Amount: currency.Amount{
					Value:    1000,
					Currency: currency.USD,
					Scale:    2,
				},
				DestinationIdentityType: "Twitter",
				DestinationIdentity:     "@elon",
				LinkedAccountTitle:      "VISA XXX123",
			},
		},
		{
			name: "success with foreign id",
			args: transactions.CreateTransactionArgs{
				WalletID:    uuid.NewString(),
				ForeignID:   uuid.NewString(),
				ForeignType: transactions.TransactionTypeOpenOutgoingPayment,
				Provider:    transactions.ProviderPaymentsEngine,
				State:       transactions.StatePending,
				Source:      "$ilp.link/alice",
				Destination: "$ilp.link/bob",
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
				Provider:    transactions.ProviderPaymentsEngine,
				State:       transactions.StatePending,
				Source:      "$ilp.link/alice",
				Destination: "$ilp.link/bob",
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
				Provider:    transactions.ProviderPaymentsEngine,
				State:       transactions.StatePending,
				Source:      "$ilp.link/alice",
				Destination: "$ilp.link/bob",
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
			walletID := uuid.NewString()

			la, err := laClient.Create(ctx, &linkedaccounts.CreateArgs{
				WalletID:   walletID,
				Name:       "test",
				Mask:       "ladida",
				Provider:   pti.ProviderName,
				ProviderID: uuid.NewString(),
				Type:       "test",
			})
			require.NoError(t, err)

			tc.args.WalletID = walletID
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
				Provider:    transactions.ProviderPaymentsEngine,
				State:       transactions.StatePending,
				Source:      "$ilp.link/alice",
				Destination: "$ilp.link/bob",
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
				Provider:    transactions.ProviderPaymentsEngine,
				State:       transactions.StatePending,
				Source:      "$ilp.link/alice",
				Destination: "$ilp.link/bob",
				Amount: currency.Amount{
					Value:    1000,
					Currency: currency.USD,
					Scale:    2,
				},
			},
		},
		{
			name: "include failed",
			len:  1,
			args: transactions.CreateTransactionArgs{
				WalletID:    uuid.NewString(),
				ForeignID:   uuid.NewString(),
				ForeignType: transactions.TransactionTypeOpenOutgoingPayment,
				Provider:    transactions.ProviderPaymentsEngine,
				State:       transactions.StateFailed,
				Source:      "$ilp.link/alice",
				Destination: "$ilp.link/bob",
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
				WalletID:                uuid.NewString(),
				ForeignID:               uuid.NewString(),
				ForeignType:             transactions.TransactionTypeOpenOutgoingPayment,
				Provider:                transactions.ProviderPaymentsEngine,
				State:                   transactions.StatePending,
				Source:                  "$ilp.link/alice",
				Destination:             "$ilp.link/bob",
				LinkedAccountTitle:      "VISA XXXX 1234",
				DestinationIdentityType: "Twitter",
				DestinationIdentity:     "@elon",
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
			walletID := uuid.NewString()

			tc.args.WalletID = walletID
			la, err := laClient.Create(ctx, &linkedaccounts.CreateArgs{
				WalletID:   walletID,
				Name:       "test",
				Mask:       "ladida",
				Provider:   pti.ProviderName,
				ProviderID: uuid.NewString(),
				Type:       "test",
			})
			require.NoError(t, err)

			tc.args.WalletID = walletID
			for i := range tc.args.Transfers {
				tc.args.Transfers[i].LinkedAccountID = la.ID
			}

			_, err = ops.CreateTransaction(ctx, b, tc.args)
			require.NoError(t, err)

			txs, err := ops.ListWithPending(ctx, b, walletID, db.Pagination{})
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
				assert.Equal(t, tc.args.DestinationIdentity, tx.DestinationIdentity)
				assert.Equal(t, tc.args.DestinationIdentityType, tx.DestinationIdentityType)
				assert.Equal(t, tc.args.LinkedAccountTitle, tx.LinkedAccountTitle)
			}
		})
	}
}

func TestListWithPendingPagination(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	dbc := db.MigrateTestDB(t, ctx)

	b := ops.NewTestBackends(t, dbc)

	pendingTxs := make([]transactions.CreateTransactionArgs, 20)
	for i := range pendingTxs {
		pendingTxs[i] = transactions.CreateTransactionArgs{
			WalletID:    uuid.NewString(),
			ForeignID:   uuid.NewString(),
			ForeignType: transactions.TransactionTypeOpenOutgoingPayment,
			Provider:    transactions.ProviderPaymentsEngine,
			State:       transactions.StatePending,
			Source:      "$ilp.link/alice",
			Destination: "$ilp.link/bob",
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

			walletID := uuid.NewString()

			for i, tx := range pendingTxs {
				tx.WalletID = walletID
				txId, err := ops.CreateTransaction(ctx, b, tx)
				require.NoError(t, err)
				if tc.start != 0 && tc.start == len(pendingTxs)-i-1 {
					tc.args.PageToken = txId
				}
			}

			txs, err := ops.ListWithPending(ctx, b, walletID, tc.args)
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
				Provider:    transactions.ProviderPaymentsEngine,
				State:       transactions.StatePending,
				Source:      "$ilp.link/alice",
				Destination: "$ilp.link/bob",
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
			walletID := uuid.NewString()

			tc.args.WalletID = walletID
			la, err := laClient.Create(ctx, &linkedaccounts.CreateArgs{
				WalletID:   walletID,
				Name:       "test",
				Mask:       "ladida",
				Provider:   pti.ProviderName,
				ProviderID: uuid.NewString(),
				Type:       "test",
			})
			require.NoError(t, err)

			tc.args.WalletID = walletID
			for i := range tc.args.Transfers {
				tc.args.Transfers[i].LinkedAccountID = la.ID
			}

			trxID, err := ops.CreateTransaction(ctx, b, tc.args)
			require.NoError(t, err)

			tx, err := ops.GetTransaction(ctx, b, walletID, trxID)
			require.NoError(t, err)
			require.Empty(t, tx.ForeignID)

			err = ops.SetTransactionForeignID(ctx, b, trxID, tc.foreignID)
			require.NoError(t, err)

			tx, err = ops.GetTransaction(ctx, b, walletID, trxID)
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
				Provider:    transactions.ProviderPaymentsEngine,
				State:       transactions.StatePending,
				Source:      "$ilp.link/alice",
				Destination: "$ilp.link/bob",
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
			walletID := uuid.NewString()

			tc.args.WalletID = walletID
			la, err := laClient.Create(ctx, &linkedaccounts.CreateArgs{
				WalletID:   walletID,
				Name:       "test",
				Mask:       "ladida",
				Provider:   pti.ProviderName,
				ProviderID: uuid.NewString(),
				Type:       "test",
			})
			require.NoError(t, err)

			tc.args.WalletID = walletID
			for i := range tc.args.Transfers {
				tc.args.Transfers[i].LinkedAccountID = la.ID
			}

			trxID, err := ops.CreateTransaction(ctx, b, tc.args)
			require.NoError(t, err)

			tx, err := ops.GetTransaction(ctx, b, walletID, trxID)
			require.NoError(t, err)
			require.Empty(t, tx.ForeignID)

			xfers, err := ops.ListTransfers(ctx, b, trxID)
			require.NoError(t, err)

			xfer := xfers[0]
			require.Empty(t, xfer.ForeignID)

			err = ops.SetTransferForeignID(ctx, b, xfer.ID, tc.foreignID)
			require.NoError(t, err)

			xfers, err = ops.ListTransfers(ctx, b, trxID)
			require.NoError(t, err)

			assert.Equal(t, xfers[0].ForeignID, tc.foreignID)
		})
	}
}

func TestSetTransactionState(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	dbc := db.MigrateTestDB(t, ctx)

	b := ops.NewTestBackends(t, dbc)
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
				Provider:    transactions.ProviderPaymentsEngine,
				State:       transactions.StatePending,
				Source:      "$ilp.link/alice",
				Destination: "$ilp.link/bob",
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
			walletID := uuid.NewString()

			tc.args.WalletID = walletID
			la, err := laClient.Create(ctx, &linkedaccounts.CreateArgs{
				WalletID:   walletID,
				Name:       "test",
				Mask:       "ladida",
				Provider:   pti.ProviderName,
				ProviderID: uuid.NewString(),
				Type:       "test",
			})
			require.NoError(t, err)

			tc.args.WalletID = walletID
			for i := range tc.args.Transfers {
				tc.args.Transfers[i].LinkedAccountID = la.ID
			}

			trxID, err := ops.CreateTransaction(ctx, b, tc.args)
			require.NoError(t, err)

			tx, err := ops.GetTransaction(ctx, b, walletID, trxID)
			require.NoError(t, err)
			require.Equal(t, tx.State, tc.args.State)

			err = ops.SetTransactionState(ctx, b, trxID, tc.state)
			require.NoError(t, err)

			tx, err = ops.GetTransaction(ctx, b, walletID, trxID)
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
				Provider:    transactions.ProviderPaymentsEngine,
				State:       transactions.StatePending,
				Source:      "$ilp.link/alice",
				Destination: "$ilp.link/bob",
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
			walletID := uuid.NewString()

			tc.args.WalletID = walletID
			la, err := laClient.Create(ctx, &linkedaccounts.CreateArgs{
				WalletID:   walletID,
				Name:       "test",
				Mask:       "ladida",
				Provider:   pti.ProviderName,
				ProviderID: uuid.NewString(),
				Type:       "test",
			})
			require.NoError(t, err)

			tc.args.WalletID = walletID
			for i := range tc.args.Transfers {
				tc.args.Transfers[i].LinkedAccountID = la.ID
			}

			trxID, err := ops.CreateTransaction(ctx, b, tc.args)
			require.NoError(t, err)

			xfers, err := ops.ListTransfers(ctx, b, trxID)
			require.NoError(t, err)
			xfer := xfers[0]
			require.Equal(t, xfer.State, tc.args.Transfers[0].State)

			err = ops.SetTransferState(ctx, b, xfer.ID, tc.state)
			require.NoError(t, err)

			xfers, err = ops.ListTransfers(ctx, b, trxID)
			require.NoError(t, err)

			assert.Equal(t, xfers[0].State, tc.state)
		})
	}
}
