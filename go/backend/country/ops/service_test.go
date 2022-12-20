package ops_test

import (
	"context"
	"testing"

	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gitlab.com/fynbos/backend/country/ops"
	"gitlab.com/fynbos/backend/db"
)

func TestGetCountries(t *testing.T) {
	b := TestBackends{
		db: db.MigrateTestDB(t, context.Background()),
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
