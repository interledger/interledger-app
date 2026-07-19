package db_test

import (
	"github.com/interledger/interledger-app/go/backend/db"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestOrderBy(t *testing.T) {
	type test struct {
		name                string
		inputOrderBy        string
		inputAllowedColumns []string
		inputTable          string
		err                 error
		expectedOrderBy     string
		expectedWhere       string
	}

	tests := []test{
		{
			name:                "default desc",
			inputOrderBy:        "name desc",
			inputAllowedColumns: []string{"name"},
			inputTable:          "contacts",
			err:                 nil,
			expectedOrderBy:     "ORDER BY name desc,id asc",
			expectedWhere:       "WHERE (name > (select name from contacts where id = $1) OR (name = (select name from contacts where id = $1) AND id > $1))",
		},
		{
			name:                "default asc if not specified",
			inputOrderBy:        "name",
			inputAllowedColumns: []string{"name"},
			inputTable:          "contacts",
			err:                 nil,
			expectedOrderBy:     "ORDER BY name asc,id asc",
			expectedWhere:       "WHERE (name > (select name from contacts where id = $1) OR (name = (select name from contacts where id = $1) AND id > $1))",
		},
		{
			name:                "error if incorrect direction",
			inputOrderBy:        "name upsidedown",
			inputAllowedColumns: []string{"name"},
			inputTable:          "contacts",
			err:                 db.OrderByDirectionViolationError,
			expectedOrderBy:     "",
			expectedWhere:       "",
		},
		{
			name:                "enforce column is allowed",
			inputOrderBy:        "name upsidedown",
			inputAllowedColumns: []string{"surname"},
			inputTable:          "contacts",
			err:                 db.OrderByColumnViolationError,
			expectedOrderBy:     "",
			expectedWhere:       "",
		},
	}

	for _, tc := range tests {
		ob, err := db.NewOrderBy(tc.inputOrderBy, tc.inputAllowedColumns, tc.inputTable)
		if tc.err != nil {
			require.ErrorIs(t, tc.err, err)
			return
		}
		if err != nil {
			t.Fatal("unexpected error", err.Error())
		}

		require.Equal(t, tc.expectedOrderBy, ob.SQLOrderBy())
		require.Equal(t, tc.expectedWhere, ob.SQLWhere("$1"))
	}
}
