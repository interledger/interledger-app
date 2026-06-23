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

func TestConfigureLedgers(t *testing.T) {
	ctx := context.Background()

	_, db := db.MigrateTestDB(t, ctx)
	b := test_utils.NewBackends(t, db)

	cases := []struct {
		name  string
		input []pacioli.ConfigureLedgerArgs
		err   error
		res   []pacioli.LedgerResult
	}{
		{
			name: "success 1",
			input: []pacioli.ConfigureLedgerArgs{
				{
					ID:    1,
					Name:  "USD_LEDGER",
					Asset: "USD",
					Scale: 2,
				},
			},
		},
		{
			name: "duplicate noop success",
			input: []pacioli.ConfigureLedgerArgs{
				{
					ID:    2,
					Name:  "ZAR_LEDGER",
					Asset: "ZAR",
					Scale: 2,
				},
				{
					ID:    2,
					Name:  "ZAR_LEDGER",
					Asset: "ZAR",
					Scale: 2,
				},
			},
		},
		{
			name: "duplicates different values",
			input: []pacioli.ConfigureLedgerArgs{
				{
					ID:    3,
					Name:  "EUR_LEDGER",
					Asset: "EUR",
					Scale: 2,
				},
				{
					ID:    3,
					Name:  "EUR_LEDGER",
					Asset: "GBP",
					Scale: 2,
				},
				{
					ID:    3,
					Name:  "EUR_LEDGER_DUPLICATE",
					Asset: "EUR",
					Scale: 2,
				},
				{
					ID:    3,
					Name:  "EUR_LEDGER",
					Asset: "EUR",
					Scale: 3,
				},
			},
			res: []pacioli.LedgerResult{
				{
					Code:  pacioli.LedgerExistsWithDifferentAsset,
					Index: 1,
				},
				{
					Code:  pacioli.LedgerExistsWithDifferentName,
					Index: 2,
				},
				{
					Code:  pacioli.LedgerExistsWithDifferentScale,
					Index: 3,
				},
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			res, err := tigerroach.ConfigureLedgers(ctx, b, tc.input)
			if tc.err != nil {
				require.ErrorIs(t, err, tc.err)
				return
			}

			require.NoError(t, err)
			require.Len(t, res, len(tc.res))

			for i, lr := range res {
				elr := tc.res[i]
				assert.Equal(t, elr.Index, lr.Index)
				assert.Equal(t, elr.Code, lr.Code)
			}
		})
	}
}
