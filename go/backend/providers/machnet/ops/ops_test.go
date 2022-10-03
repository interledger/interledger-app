package ops_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gitlab.com/fynbos/backend/providers/machnet"
	"gitlab.com/fynbos/backend/providers/machnet/ops"
	test_utils "gitlab.com/fynbos/backend/utils"
)

func TestCreateAndGetUser(t *testing.T) {
	testDB := test_utils.MigrateCockroachDB(t, context.Background())
	b := backends{db: testDB}
	walletID := uuid.NewString()
	_, err := testDB.Exec(
		"INSERT INTO wallets (id, name) VALUES ($1, $2);",
		walletID,
		"test",
	)
	require.NoError(t, err)

	args := machnet.CreateArgs{
		WalletID:   walletID,
		ExternalID: uuid.NewString(),
	}
	user, err := ops.CreateUser(context.Background(), b, args)
	require.NoError(t, err)
	require.Equal(t, args.ExternalID, user.ID)

	freshUser, err := ops.GetUser(context.Background(), b, walletID)
	require.NoError(t, err)
	require.Equal(t, args.ExternalID, freshUser.ID)

	noUser, err := ops.GetUser(context.Background(), b, uuid.NewString())
	assert.Nil(t, noUser)
	assert.ErrorIs(t, err, machnet.ErrNotFound)
}

type backends struct {
	db *sqlx.DB
}

func (b backends) DB() *sqlx.DB {
	return b.db
}
