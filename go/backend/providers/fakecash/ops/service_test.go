package ops_test

import (
	"context"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gitlab.com/fynbos/backend/db"
	"gitlab.com/fynbos/backend/providers/fakecash"
	"gitlab.com/fynbos/backend/providers/fakecash/ops"
	"gitlab.com/fynbos/pacioli"
	pacioli_mock "gitlab.com/fynbos/pacioli/client/mock"
)

func TestCreateAndGet(t *testing.T) {
	b := newBackends(t)

	accountID := uuid.NewString()
	b.pacioli.EXPECT().ConfigureAccounts(gomock.Any(), gomock.Any()).Return(
		[]pacioli.AccountResult{}, nil,
	).AnyTimes()

	account, err := ops.Create(context.Background(), b, fakecash.CreateArgs{ID: accountID})
	require.NoError(t, err)
	require.Equal(t, accountID, account.ID)
	require.Equal(t, uint64(0), account.AvailableBalance)

	b.pacioli.EXPECT().GetAccounts(gomock.Any(), []string{account.ID}).Return(
		[]pacioli.Account{
			{ID: account.ID},
		}, nil,
	).AnyTimes()
	freshAccount, err := ops.Get(context.Background(), b, account.ID)
	require.NoError(t, err)
	assert.Equal(t, accountID, freshAccount.ID)
	assert.Equal(t, uint64(0), freshAccount.AvailableBalance)
}

type backends struct {
	pacioli  *pacioli_mock.MockClient
	ledgerID uint32
	db       *sqlx.DB
}

func (b backends) Pacioli() pacioli.Client {
	return b.pacioli
}

func (b backends) LedgerID() uint32 {
	return b.ledgerID
}

func (b backends) DB() *sqlx.DB {
	return b.db
}

func newBackends(t *testing.T) *backends {
	ctrl := gomock.NewController(t)
	t.Cleanup(func() {
		ctrl.Finish()
	})

	return &backends{
		db:       db.MigrateTestDB(t, context.Background()),
		ledgerID: uint32(10),
		pacioli:  pacioli_mock.NewMockClient(ctrl),
	}
}
