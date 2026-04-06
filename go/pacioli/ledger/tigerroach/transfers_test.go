package tigerroach_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gitlab.com/fynbos/pacioli"
	"gitlab.com/fynbos/pacioli/db"
	"gitlab.com/fynbos/pacioli/ledger/tigerroach"
	test_utils "gitlab.com/fynbos/pacioli/utils"
)

func TestCreateTransfers(t *testing.T) {
	ctx := context.Background()

	_, db := db.MigrateTestDB(t, ctx)
	b := test_utils.NewBackends(t, db)

	// Configure Ledger
	lr, err := tigerroach.ConfigureLedgers(ctx, b, []pacioli.ConfigureLedgerArgs{
		{
			ID:    1,
			Name:  "TestLedgerUSD",
			Asset: "USD",
			Scale: 2,
		},
		{
			ID:    2,
			Name:  "TestLedgerZAR",
			Asset: "ZAR",
			Scale: 2,
		},
	})
	require.NoError(t, err)
	require.Empty(t, lr)

	cases := []struct {
		name  string
		input []pacioli.CreateTransferArgs
		err   error
		res   []pacioli.TransferResult
	}{
		{
			name: "success posted",
			input: []pacioli.CreateTransferArgs{
				{
					ID:              "62eb03aa-2e73-464a-be1e-547429ddc86a",
					Amount:          1000,
					DebitAccountID:  "aace20cf-177d-418d-93bd-f6d9cb0d49d1",
					CreditAccountID: "e139db65-fed1-4e1b-a15b-b1c98d718e39",
					Code:            1,
					Ledger:          1,
				},
			},
		},
		{
			name: "success pending",
			input: []pacioli.CreateTransferArgs{
				{
					ID:              "a0d03fc7-f7a8-486b-8b4a-2a5786e02f76",
					Amount:          1000,
					DebitAccountID:  "b9877da9-a44c-413a-b0c2-4dbfac4e1dcd",
					CreditAccountID: "aad81360-b31a-4d70-ab28-acd8fb0e3d26",
					Code:            1,
					Ledger:          1,
					Pending:         true,
					Timeout:         uint64(time.Minute * 20),
				},
			},
		},
		{
			name: "success duplicate",
			input: []pacioli.CreateTransferArgs{
				{
					ID:              "09c135b0-a648-4298-9bf9-509890168968",
					Amount:          1000,
					DebitAccountID:  "dae2b2a6-3a1f-4818-af68-85dd7c34ed49",
					CreditAccountID: "0c895c5b-a369-41a8-8071-e69394b2f269",
					Code:            1,
					Ledger:          1,
				},
				{
					ID:              "09c135b0-a648-4298-9bf9-509890168968",
					Amount:          1000,
					DebitAccountID:  "dae2b2a6-3a1f-4818-af68-85dd7c34ed49",
					CreditAccountID: "0c895c5b-a369-41a8-8071-e69394b2f269",
					Code:            1,
					Ledger:          1,
				},
			},
		},
		{
			name: "success duplicate pending",
			input: []pacioli.CreateTransferArgs{
				{
					ID:              "7f000ca0-c6e8-4e9b-993c-a27aca075b97",
					Amount:          1000,
					DebitAccountID:  "d7374522-c847-4bec-856b-f607743bb6d3",
					CreditAccountID: "35993786-91e0-49bd-8699-dd0c4870dd30",
					Code:            1,
					Ledger:          1,
					Pending:         true,
					Timeout:         uint64(time.Minute * 20),
				},
				{
					ID:              "7f000ca0-c6e8-4e9b-993c-a27aca075b97",
					Amount:          1000,
					DebitAccountID:  "d7374522-c847-4bec-856b-f607743bb6d3",
					CreditAccountID: "35993786-91e0-49bd-8699-dd0c4870dd30",
					Code:            1,
					Ledger:          1,
					Pending:         true,
					Timeout:         uint64(time.Minute * 20),
				},
			},
		},
		{
			name: "existing non match errors",
			input: []pacioli.CreateTransferArgs{
				{
					// Success
					ID:              "db3b784d-fe3a-4604-a127-1086f6f3dbf9",
					Amount:          1000,
					DebitAccountID:  "a9f5ddd2-995f-47af-9c2a-6c7ef2084d22",
					CreditAccountID: "d85b0d50-7058-4ba8-af2e-b57ae6456c5b",
					Code:            1,
					Ledger:          1,
				},
				{
					// Different Amount
					ID:              "db3b784d-fe3a-4604-a127-1086f6f3dbf9",
					Amount:          4000,
					DebitAccountID:  "a9f5ddd2-995f-47af-9c2a-6c7ef2084d22",
					CreditAccountID: "d85b0d50-7058-4ba8-af2e-b57ae6456c5b",
					Code:            1,
					Ledger:          1,
				},
				{
					// Different Debit account
					ID:              "db3b784d-fe3a-4604-a127-1086f6f3dbf9",
					Amount:          1000,
					DebitAccountID:  "7857b38c-1acb-4bf0-804d-1878eb7bfd56",
					CreditAccountID: "d85b0d50-7058-4ba8-af2e-b57ae6456c5b",
					Code:            1,
					Ledger:          1,
				},
				{
					// Different Credit Account
					ID:              "db3b784d-fe3a-4604-a127-1086f6f3dbf9",
					Amount:          1000,
					DebitAccountID:  "a9f5ddd2-995f-47af-9c2a-6c7ef2084d22",
					CreditAccountID: "95943c31-2f20-4744-800e-d9b8b5199918",
					Code:            1,
					Ledger:          1,
				},
				{
					// Different Code
					ID:              "db3b784d-fe3a-4604-a127-1086f6f3dbf9",
					Amount:          1000,
					DebitAccountID:  "a9f5ddd2-995f-47af-9c2a-6c7ef2084d22",
					CreditAccountID: "d85b0d50-7058-4ba8-af2e-b57ae6456c5b",
					Code:            2,
					Ledger:          1,
				},
			},
			res: []pacioli.TransferResult{
				{
					Index: 1,
					Code:  pacioli.TransferExistsWithDifferentAmount,
				},
				{
					Index: 2,
					Code:  pacioli.TransferExistsWithDifferentDebitAccountId,
				},
				{
					Index: 3,
					Code:  pacioli.TransferExistsWithDifferentCreditAccountId,
				},
				{
					Index: 4,
					Code:  pacioli.TransferExistsWithDifferentCode,
				},
			},
		},
		{
			name: "validation errors",
			input: []pacioli.CreateTransferArgs{
				{
					ID:              "fc35bd84-9a6d-4536-988e-5add749cf16b",
					Amount:          1000,
					DebitAccountID:  "403a4d1d-3780-4199-a5c9-7251fc9b3fda",
					CreditAccountID: "403a4d1d-3780-4199-a5c9-7251fc9b3fda",
					Code:            1,
					Ledger:          1,
				},
				{
					ID:              "e78c6564-7562-47c3-9452-0f42a4be177c",
					Amount:          1000,
					DebitAccountID:  "acc817ba-9482-421e-8d32-ac44d7cdce41",
					CreditAccountID: "b0288fe7-9198-43f0-af9e-1e54ec2ab16a",
					Code:            1,
					Ledger:          1,
					Pending:         true,
				},
				{
					ID:              "e78c6564-7562-47c3-9452-0f42a4be177c",
					Amount:          1000,
					DebitAccountID:  "acc817ba-9482-421e-8d32-ac44d7cdce41",
					CreditAccountID: "b0288fe7-9198-43f0-af9e-1e54ec2ab16a",
					Code:            1,
					Ledger:          2,
				},
			},
			res: []pacioli.TransferResult{
				{
					Index: 0,
					Code:  pacioli.TransferAccountsMustBeDifferent,
				},
				{
					Index: 1,
					Code:  pacioli.TransferPendingTransferMustTimeout,
				},
				{
					Index: 2,
					Code:  pacioli.TransferTransferMustHaveTheSameLedgerAsAccounts,
				},
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {

			// Configure the accounts
			for _, args := range tc.input {
				ar, err := tigerroach.ConfigureAccounts(ctx, b, []pacioli.ConfigureAccountArgs{
					{
						ID:       args.CreditAccountID,
						LedgerID: 1,
						Code:     1,
					},
					{
						ID:       args.DebitAccountID,
						LedgerID: 1,
						Code:     1,
					},
				})
				require.NoError(t, err)
				require.Empty(t, ar)
			}

			res, err := tigerroach.CreateTransfers(ctx, b, tc.input)
			if tc.err != nil {
				require.ErrorIs(t, err, tc.err)
				return
			}

			require.NoError(t, err)
			require.Len(t, res, len(tc.res))

			for i, tr := range res {
				etr := tc.res[i]
				assert.Equal(t, etr.Index, tr.Index)
				assert.Equal(t, etr.Code, tr.Code)
			}

			for i, args := range tc.input {
				var skipValidation bool
				for _, ee := range tc.res {
					if ee.Index == uint32(i) {
						skipValidation = true
						break
					}
				}
				if skipValidation {
					continue
				}

				tr, err := tigerroach.GetTransfer(ctx, b, args.ID)
				assert.NoError(t, err)
				assert.Equal(t, tr.ID, args.ID)

				da, err := tigerroach.GetAccount(ctx, b, args.DebitAccountID)
				assert.NoError(t, err)
				if args.Pending {
					assert.Equal(t, args.Amount, int64(da.DebitsPending))
				} else {
					assert.Equal(t, args.Amount, int64(da.DebitsPosted))
				}

				ca, err := tigerroach.GetAccount(ctx, b, args.CreditAccountID)
				assert.NoError(t, err)
				if args.Pending {
					assert.Equal(t, args.Amount, int64(ca.CreditsPending))
				} else {
					assert.Equal(t, args.Amount, int64(ca.CreditsPosted))
				}
			}
		})
	}
}

func TestPostTransfers(t *testing.T) {
	ctx := context.Background()

	_, db := db.MigrateTestDB(t, ctx)
	b := test_utils.NewBackends(t, db)

	// Configure Ledger
	lr, err := tigerroach.ConfigureLedgers(ctx, b, []pacioli.ConfigureLedgerArgs{
		{
			ID:    1,
			Name:  "TestLedgerUSD",
			Asset: "USD",
			Scale: 2,
		},
	})
	require.NoError(t, err)
	require.Empty(t, lr)
	cases := []struct {
		name  string
		input []pacioli.CreateTransferArgs
		err   error
		res   []pacioli.TransferResult
	}{
		{
			name: "success single",
			input: []pacioli.CreateTransferArgs{
				{
					ID:              "7f000ca0-c6e8-4e9b-993c-a27aca075b97",
					Amount:          1000,
					DebitAccountID:  "d7374522-c847-4bec-856b-f607743bb6d3",
					CreditAccountID: "35993786-91e0-49bd-8699-dd0c4870dd30",
					Code:            1,
					Ledger:          1,
					Pending:         true,
					Timeout:         uint64(time.Minute * 20),
				},
			},
		},
		{
			name: "already posted",
			input: []pacioli.CreateTransferArgs{
				{
					ID:              "7ccaf0be-d114-4e66-b2b9-21389f5ca996",
					Amount:          1000,
					DebitAccountID:  "f35459cd-1ea0-4402-8462-8b1a15fa8841",
					CreditAccountID: "025f782d-6389-4307-a08d-43fee5b1ec34",
					Code:            1,
					Ledger:          1,
					Pending:         true,
					Timeout:         uint64(time.Minute * 20),
				},
				{
					ID:              "7ccaf0be-d114-4e66-b2b9-21389f5ca996",
					Amount:          1000,
					DebitAccountID:  "f35459cd-1ea0-4402-8462-8b1a15fa8841",
					CreditAccountID: "025f782d-6389-4307-a08d-43fee5b1ec34",
					Code:            1,
					Ledger:          1,
					Pending:         true,
					Timeout:         uint64(time.Minute * 20),
				},
			},
			res: []pacioli.TransferResult{
				{
					Index: 1,
					Code:  pacioli.TransferPendingTransferAlreadyPosted,
				},
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			// Configure the accounts
			var tids []string
			for _, args := range tc.input {
				ar, err := tigerroach.ConfigureAccounts(ctx, b, []pacioli.ConfigureAccountArgs{
					{
						ID:       args.CreditAccountID,
						LedgerID: 1,
						Code:     1,
					},
					{
						ID:       args.DebitAccountID,
						LedgerID: 1,
						Code:     1,
					},
				})
				require.NoError(t, err)
				require.Empty(t, ar)
				tids = append(tids, args.ID)
			}

			// Create the transfers
			tr, err := tigerroach.CreateTransfers(ctx, b, tc.input)
			require.NoError(t, err)
			require.Empty(t, tr)

			// Post the transfers
			_, prl, err := tigerroach.PostTransfers(ctx, b, tids)
			if tc.err != nil {
				require.ErrorIs(t, err, tc.err)
				return
			}

			require.NoError(t, err)
			require.Len(t, prl, len(tc.res))

			for i, pr := range prl {
				etr := tc.res[i]
				assert.Equal(t, etr.Index, pr.Index)
				assert.Equal(t, etr.Code, pr.Code)
			}

			// Lookup the transaction/accounts and check their state and debit and credit values
			for i, args := range tc.input {
				var skipValidation bool
				for _, ee := range tc.res {
					if ee.Index == uint32(i) {
						skipValidation = true
						break
					}
				}
				if skipValidation {
					continue
				}

				tr, err := tigerroach.GetTransfer(ctx, b, args.ID)
				assert.NoError(t, err)
				assert.Equal(t, tr.ID, args.ID)
				assert.NotEqual(t, pacioli.TransferStatePending, tr.State)

				da, err := tigerroach.GetAccount(ctx, b, args.DebitAccountID)
				assert.NoError(t, err)
				assert.Zero(t, da.DebitsPending)
				assert.Equal(t, args.Amount, int64(da.DebitsPosted))

				ca, err := tigerroach.GetAccount(ctx, b, args.CreditAccountID)
				assert.NoError(t, err)
				assert.Zero(t, ca.CreditsPending)
				assert.Equal(t, args.Amount, int64(ca.CreditsPosted))
			}
		})
	}
}

func TestVoidTransfers(t *testing.T) {
	ctx := context.Background()

	_, db := db.MigrateTestDB(t, ctx)
	b := test_utils.NewBackends(t, db)

	// Configure Ledger
	lr, err := tigerroach.ConfigureLedgers(ctx, b, []pacioli.ConfigureLedgerArgs{
		{
			ID:    1,
			Name:  "TestLedgerUSD",
			Asset: "USD",
			Scale: 2,
		},
	})
	require.NoError(t, err)
	require.Empty(t, lr)
	cases := []struct {
		name  string
		input []pacioli.CreateTransferArgs
		err   error
		res   []pacioli.TransferResult
	}{
		{
			name: "success single",
			input: []pacioli.CreateTransferArgs{
				{
					ID:              "7f000ca0-c6e8-4e9b-993c-a27aca075b97",
					Amount:          1000,
					DebitAccountID:  "d7374522-c847-4bec-856b-f607743bb6d3",
					CreditAccountID: "35993786-91e0-49bd-8699-dd0c4870dd30",
					Code:            1,
					Ledger:          1,
					Pending:         true,
					Timeout:         uint64(time.Minute * 20),
				},
			},
		},
		{
			name: "already voided",
			input: []pacioli.CreateTransferArgs{
				{
					ID:              "7ccaf0be-d114-4e66-b2b9-21389f5ca996",
					Amount:          1000,
					DebitAccountID:  "f35459cd-1ea0-4402-8462-8b1a15fa8841",
					CreditAccountID: "025f782d-6389-4307-a08d-43fee5b1ec34",
					Code:            1,
					Ledger:          1,
					Pending:         true,
					Timeout:         uint64(time.Minute * 20),
				},
				{
					ID:              "7ccaf0be-d114-4e66-b2b9-21389f5ca996",
					Amount:          1000,
					DebitAccountID:  "f35459cd-1ea0-4402-8462-8b1a15fa8841",
					CreditAccountID: "025f782d-6389-4307-a08d-43fee5b1ec34",
					Code:            1,
					Ledger:          1,
					Pending:         true,
					Timeout:         uint64(time.Minute * 20),
				},
			},
			res: []pacioli.TransferResult{
				{
					Index: 1,
					Code:  pacioli.TransferPendingTransferAlreadyVoided,
				},
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			// Configure the accounts
			var tids []string
			for _, args := range tc.input {
				ar, err := tigerroach.ConfigureAccounts(ctx, b, []pacioli.ConfigureAccountArgs{
					{
						ID:       args.CreditAccountID,
						LedgerID: 1,
						Code:     1,
					},
					{
						ID:       args.DebitAccountID,
						LedgerID: 1,
						Code:     1,
					},
				})
				require.NoError(t, err)
				require.Empty(t, ar)
				tids = append(tids, args.ID)
			}

			// Create the transfers
			tr, err := tigerroach.CreateTransfers(ctx, b, tc.input)
			require.NoError(t, err)
			require.Empty(t, tr)

			// Void the transfers
			prl, err := tigerroach.VoidTransfers(ctx, b, tids)
			if tc.err != nil {
				require.ErrorIs(t, err, tc.err)
				return
			}
			require.NoError(t, err)
			require.Len(t, prl, len(tc.res))

			for i, pr := range prl {
				etr := tc.res[i]
				assert.Equal(t, etr.Index, pr.Index)
				assert.Equal(t, etr.Code, pr.Code)
			}

			// Lookup the transaction/accounts and check their state and debit and credit values
			for i, args := range tc.input {
				var skipValidation bool
				for _, ee := range tc.res {
					if ee.Index == uint32(i) {
						skipValidation = true
						break
					}
				}
				if skipValidation {
					continue
				}

				tr, err := tigerroach.GetTransfer(ctx, b, args.ID)
				assert.NoError(t, err)
				assert.Equal(t, tr.ID, args.ID)
				assert.NotEqual(t, pacioli.TransferStatePending, tr.State)

				da, err := tigerroach.GetAccount(ctx, b, args.DebitAccountID)
				assert.NoError(t, err)
				assert.Zero(t, da.DebitsPending)

				ca, err := tigerroach.GetAccount(ctx, b, args.CreditAccountID)
				assert.NoError(t, err)
				assert.Zero(t, ca.CreditsPending)
			}
		})
	}
}

func TestTimeoutTransfers(t *testing.T) {
	ctx := context.Background()

	_, db := db.MigrateTestDB(t, ctx)
	b := test_utils.NewBackends(t, db)

	// Configure Ledger
	lr, err := tigerroach.ConfigureLedgers(ctx, b, []pacioli.ConfigureLedgerArgs{
		{
			ID:    1,
			Name:  "TestLedgerUSD",
			Asset: "USD",
			Scale: 2,
		},
	})
	require.NoError(t, err)
	require.Empty(t, lr)
	cases := []struct {
		name  string
		input []pacioli.CreateTransferArgs
		err   error
	}{
		{
			name: "success single",
			input: []pacioli.CreateTransferArgs{
				{
					ID:              "7f000ca0-c6e8-4e9b-993c-a27aca075b97",
					Amount:          1000,
					DebitAccountID:  "d7374522-c847-4bec-856b-f607743bb6d3",
					CreditAccountID: "35993786-91e0-49bd-8699-dd0c4870dd30",
					Code:            1,
					Ledger:          1,
					Pending:         true,
					Timeout:         uint64(time.Minute * 20),
				},
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			// Configure the accounts
			var tids []string
			for _, args := range tc.input {
				ar, err := tigerroach.ConfigureAccounts(ctx, b, []pacioli.ConfigureAccountArgs{
					{
						ID:       args.CreditAccountID,
						LedgerID: 1,
						Code:     1,
					},
					{
						ID:       args.DebitAccountID,
						LedgerID: 1,
						Code:     1,
					},
				})
				require.NoError(t, err)
				require.Empty(t, ar)
				tids = append(tids, args.ID)
			}

			// Create the transfers
			tr, err := tigerroach.CreateTransfers(ctx, b, tc.input)
			require.NoError(t, err)
			require.Empty(t, tr)

			// Lookup the transaction/accounts and check their state and debit and credit values
			for _, args := range tc.input {

				tr, err := tigerroach.GetTransfer(ctx, b, args.ID)
				assert.NoError(t, err)
				assert.Equal(t, tr.ID, args.ID)
				assert.Equal(t, pacioli.TransferStatePending, tr.State)

				da, err := tigerroach.GetAccount(ctx, b, args.DebitAccountID)
				assert.NoError(t, err)
				assert.Greater(t, da.DebitsPending, uint64(0))

				ca, err := tigerroach.GetAccount(ctx, b, args.CreditAccountID)
				assert.NoError(t, err)
				assert.Greater(t, ca.CreditsPending, uint64(0))
			}

			// Update the transfers timeouts
			_, err = b.DB().ExecContext(ctx, "update ledger_transfers set timeout_at=$1", time.Now().UTC().Add(-time.Hour))
			require.NoError(t, err)

			_, err = tigerroach.TryTimeoutTransfers(ctx, b, tids)
			if tc.err != nil {
				require.ErrorIs(t, tc.err, err)
				return
			}

			// Lookup the transaction/accounts and check their state and debit and credit values
			for _, args := range tc.input {

				tr, err := tigerroach.GetTransfer(ctx, b, args.ID)
				assert.NoError(t, err)
				assert.Equal(t, tr.ID, args.ID)
				assert.NotEqual(t, pacioli.TransferStatePending, tr.State)

				da, err := tigerroach.GetAccount(ctx, b, args.DebitAccountID)
				assert.NoError(t, err)
				assert.Zero(t, da.DebitsPending)

				ca, err := tigerroach.GetAccount(ctx, b, args.CreditAccountID)
				assert.NoError(t, err)
				assert.Zero(t, ca.CreditsPending)
			}
		})
	}
}
