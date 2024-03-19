package ops_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gitlab.com/fynbos/backend/db"
	"gitlab.com/fynbos/backend/providers/gatehub/ops"
	"gitlab.com/fynbos/backend/user"
	user_mock "gitlab.com/fynbos/backend/user/client/mock"
)

func TestSaveUser(t *testing.T) {
	b := Backends{
		db:    db.MigrateTestDB(t, context.Background()),
		users: user_mock.NewMock(),
	}

	walletID := uuid.NewString()
	externalID := uuid.NewString()
	err := ops.SaveUser(context.Background(), b, walletID, externalID)
	require.NoError(t, err)

	var result string
	err = b.db.GetContext(context.Background(), &result, "SELECT external_id FROM gatehub_users WHERE wallet_id=$1;", walletID)
	require.NoError(t, err)

	assert.Equal(t, externalID, result)
}

type Backends struct {
	db    *sqlx.DB
	users *user_mock.MockClient
}

func (b Backends) Users() user.Client {
	return b.users
}

func (b Backends) DB() *sqlx.DB {
	return b.db
}
