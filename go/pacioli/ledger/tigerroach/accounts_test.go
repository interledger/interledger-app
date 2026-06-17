package tigerroach_test

import (
	"context"
	"testing"

	"github.com/interledger/interledger-app/go/pacioli"
	"github.com/interledger/interledger-app/go/pacioli/db"
	"github.com/interledger/interledger-app/go/pacioli/ledger/tigerroach"
	test_utils "github.com/interledger/interledger-app/go/pacioli/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConfigureAccounts(t *testing.T) {
	ctx := context.Background()

	_, db := db.MigrateTestDB(t, ctx)
	b := test_utils.NewBackends(t, db)

	// Configure Ledger
	lr, err := tigerroach.ConfigureLedgers(ctx, b, []pacioli.ConfigureLedgerArgs{
		{
			ID:    1,
			Name:  "TestLedger",
			Asset: "USD",
			Scale: 2,
		},
		{
			ID:    2,
			Name:  "TestLedger2",
			Asset: "ZAR",
			Scale: 2,
		},
	})
	require.NoError(t, err)
	require.Empty(t, lr)

	cases := []struct {
		name  string
		input []pacioli.ConfigureAccountArgs
		err   error
		res   []pacioli.AccountResult
	}{
		{
			name: "success 1 account",
			input: []pacioli.ConfigureAccountArgs{
				{
					ID:       "e139db65-fed1-4e1b-a15b-b1c98d018e39",
					LedgerID: 1,
					Code:     1,
				},
			},
		},
		{
			name: "unknown ledger",
			input: []pacioli.ConfigureAccountArgs{
				{
					ID:       "e139db65-fed1-4e1b-a15b-b1c98d718e39",
					LedgerID: 404,
					Code:     1,
				},
			},
			res: []pacioli.AccountResult{
				{
					Index: 0,
					Code:  pacioli.AccountLedgerDoesNotExist,
				},
			},
		},
		{
			name: "mutually exclusive flags",
			input: []pacioli.ConfigureAccountArgs{
				{
					ID:                         "aace20cf-177d-418d-93bd-f6d9cb0d49d1",
					LedgerID:                   1,
					Code:                       1,
					DebitsMustNotExceedCredits: true,
					CreditsMustNotExceedDebits: true,
				},
			},
			res: []pacioli.AccountResult{
				{
					Index: 0,
					Code:  pacioli.AccountMutuallyExclusiveFlags,
				},
			},
		},
		{
			name: "duplicates different values",
			input: []pacioli.ConfigureAccountArgs{
				{
					ID:       "1d46eea1-d0b8-40fe-9a6b-11650e7e2d62",
					LedgerID: 1,
					Code:     1,
				},
				{
					ID:       "1d46eea1-d0b8-40fe-9a6b-11650e7e2d62",
					LedgerID: 1,
					Code:     2,
				},
				{
					ID:       "1d46eea1-d0b8-40fe-9a6b-11650e7e2d62",
					LedgerID: 2,
					Code:     1,
				},
				{
					ID:                         "1d46eea1-d0b8-40fe-9a6b-11650e7e2d62",
					LedgerID:                   1,
					Code:                       1,
					DebitsMustNotExceedCredits: true,
				},
			},
			res: []pacioli.AccountResult{
				{
					Index: 1,
					Code:  pacioli.AccountExistsWithDifferentCode,
				},
				{
					Index: 2,
					Code:  pacioli.AccountExistsWithDifferentLedger,
				},
				{
					Index: 3,
					Code:  pacioli.AccountExistsWithDifferentFlags,
				},
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			res, err := tigerroach.ConfigureAccounts(ctx, b, tc.input)
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
		})
	}
}
