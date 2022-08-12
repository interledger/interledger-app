package tigerroach_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gitlab.com/fynbos/pacioli"
	"gitlab.com/fynbos/pacioli/ledger/tigerroach"
	test_utils "gitlab.com/fynbos/pacioli/utils"
)

func TestCreateTransfers(t *testing.T) {
	ctx := context.Background()

	_, db := test_utils.MigrateCockroachDB(t, ctx)
	b := test_utils.NewBackends(t, db, nil)

	cases := []struct {
		name  string
		input []pacioli.CreateTransferArgs
		err   error
		res   []pacioli.TransferResult
	}{
		{},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			res, err := tigerroach.CreateTransfers(ctx, b, tc.input)
			if tc.err != nil {
				require.ErrorIs(t, err, tc.err)
				return
			}

			require.NoError(t, err)
			require.Len(t, res, len(res))

			for i, tr := range res {
				etr := tc.res[i]
				assert.Equal(t, etr.Index, tr.Index)
				assert.Equal(t, etr.Code, tr.Code)
			}
		})
	}
}

/*
func setupTestAccounts(_ *testing.T, db *sqlx.DB) error {
	stmt, err := db.Prepare("INSERT INTO ledger_accounts VALUES (id, ledger_id, code)")
}*/
