package ops_test

import (
	"context"
	"testing"

	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gitlab.com/fynbos/backend/country/ops"
	test_utils "gitlab.com/fynbos/backend/utils"
)

func TestGetCountries(t *testing.T) {
	b := TestBackends{
		db: test_utils.MigrateCockroachDB(t, context.Background()),
	}

	countries, err := ops.GetAll(context.Background(), b)
	require.NoError(t, err)

	assert.Len(t, countries, 249)
}

type TestBackends struct {
	db *sqlx.DB
}

func (b TestBackends) DB() *sqlx.DB {
	return b.db
}
