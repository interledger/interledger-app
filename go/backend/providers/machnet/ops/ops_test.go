package ops_test

import (
	"context"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gitlab.com/fynbos/backend/providers/machnet"
	"gitlab.com/fynbos/backend/providers/machnet/external"
	external_client "gitlab.com/fynbos/backend/providers/machnet/external/client/inmemory"
	"gitlab.com/fynbos/backend/providers/machnet/ops"
	test_utils "gitlab.com/fynbos/backend/utils"
)

func TestCreateAndGetUser(t *testing.T) {
	t.Parallel()
	b := backends{
		db: test_utils.MigrateCockroachDB(t, context.Background()),
	}
	walletID := uuid.NewString()
	_, err := b.DB().Exec(
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

func TestGetWidgetToken(t *testing.T) {
	t.Parallel()
	b := backends{
		db:       test_utils.MigrateCockroachDB(t, context.Background()),
		external: external_client.New(),
	}
	walletID := uuid.NewString()
	_, err := b.DB().Exec(
		"INSERT INTO wallets (id, name) VALUES ($1, $2);",
		walletID,
		"test",
	)
	require.NoError(t, err)
	externalUser, err := b.External().RegisterUser(context.Background(), external.User{
		Type: external.SendUser,
	})
	require.NoError(t, err)
	user, err := ops.CreateUser(context.Background(), b, machnet.CreateArgs{
		WalletID:   walletID,
		ExternalID: externalUser.ID,
	})
	require.NoError(t, err)

	return user
}

func NewTestBackends(t *testing.T) backends {
	ctrl := gomock.NewController(t)
	return backends{
		db:       test_utils.MigrateCockroachDB(t, context.Background()),
		external: external_client.New(),
	}
}

type backends struct {
	db       *sqlx.DB
	external external.Client
}

func (b backends) DB() *sqlx.DB {
	return b.db
}

func (b backends) External() external.Client {
	return b.external
}
