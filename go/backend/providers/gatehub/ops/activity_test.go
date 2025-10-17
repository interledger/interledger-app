package ops_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gitlab.com/fynbos/backend/db"
	"gitlab.com/fynbos/backend/email"
	"gitlab.com/fynbos/backend/kyc"
	"gitlab.com/fynbos/backend/linkedaccounts"
	la_mock "gitlab.com/fynbos/backend/linkedaccounts/client/mock"
	"gitlab.com/fynbos/backend/notify"
	"gitlab.com/fynbos/backend/payments"
	"gitlab.com/fynbos/backend/providers/gatehub/ops"
	"gitlab.com/fynbos/backend/transactions"
	"gitlab.com/fynbos/backend/user"
	user_mock "gitlab.com/fynbos/backend/user/client/mock"
	"gitlab.com/fynbos/backend/wallets"
	wallet_mock "gitlab.com/fynbos/backend/wallets/client/mock"
	"gitlab.com/fynbos/pacioli"
	temporal "go.temporal.io/sdk/client"
)

func TestSaveUser(t *testing.T) {
	b := Backends{
		db:    db.MigrateTestDB(t, context.Background()),
		users: user_mock.NewMock(),
	}
	a := ops.NewActivity(b)

	walletID := uuid.NewString()
	externalID := uuid.NewString()
	err := a.SaveGatehubUser(context.Background(), walletID, externalID)
	require.NoError(t, err)

	var result string
	err = b.db.GetContext(context.Background(), &result, "SELECT external_id FROM gatehub_users WHERE wallet_id=$1;", walletID)
	require.NoError(t, err)

	err = a.SaveGatehubUser(context.Background(), walletID, externalID)
	require.NoError(t, err)

	assert.Equal(t, externalID, result)
}

func TestLinkGatehubUserToGateway(t *testing.T) {
	b := Backends{
		db:    db.MigrateTestDB(t, context.Background()),
		users: user_mock.NewMock(),
	}
	a := ops.NewActivity(b)

	walletID := uuid.NewString()
	externalID := uuid.NewString()
	err := a.SaveGatehubUser(context.Background(), walletID, externalID)
	require.NoError(t, err)

	return
}

type Backends struct {
	db    *sqlx.DB
	users *user_mock.MockClient
	la    *la_mock.MockClient
	wc    *wallet_mock.MockClient
}

func (b Backends) Payments() payments.Client {
	return nil
}

func (b Backends) Users() user.Client {
	return b.users
}

func (b Backends) DB() *sqlx.DB {
	return b.db
}

func (b Backends) Temporal() temporal.Client {
	return nil
}

func (b Backends) LinkedAccounts() linkedaccounts.Client {
	return b.la
}

func (b Backends) Wallets() wallets.Client {
	return b.wc
}

func (b Backends) Pacioli() pacioli.Client {
	return nil
}

func (b Backends) KYC() kyc.Client {
	return nil
}

func (b Backends) Transactions() transactions.Client {
	return nil
}

func (b Backends) Email() email.Client {
	return nil
}

func (b Backends) LinkGatehubUserToGateway() transactions.Client {
	return nil
}

func (b Backends) Notify() notify.Client {
	return nil
}
