package ops_test

import (
	"context"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/jmoiron/sqlx"
	"gitlab.com/fynbos/backend/db"
	"gitlab.com/fynbos/backend/kyc"
	kyc_mock "gitlab.com/fynbos/backend/kyc/client/mock"
	"gitlab.com/fynbos/backend/linkedaccounts"
	la_mock "gitlab.com/fynbos/backend/linkedaccounts/client/mock"
	"gitlab.com/fynbos/backend/payments"
	payments_mock "gitlab.com/fynbos/backend/payments/client/mock"
	temporal_mock "gitlab.com/fynbos/backend/temporal/mock"
	"gitlab.com/fynbos/backend/transactions"
	transactions_mock "gitlab.com/fynbos/backend/transactions/client/mock"
	"gitlab.com/fynbos/backend/user"
	user_mock "gitlab.com/fynbos/backend/user/client/mock"
	"gitlab.com/fynbos/backend/wallets"
	wallets_mock "gitlab.com/fynbos/backend/wallets/client/mock"
	"gitlab.com/fynbos/pacioli"
	temporal "go.temporal.io/sdk/client"
)

type Backends struct {
	db    *sqlx.DB
	kyc   *kyc_mock.MockClient
	la    *la_mock.MockClient
	users *user_mock.MockClient
	pc    *payments_mock.MockClient
	tp    *temporal_mock.MockClient
	txc   *transactions_mock.MockClient
	wc    *wallets_mock.MockClient
}

func (b Backends) Payments() payments.Client {
	return b.pc
}

func (b Backends) DB() *sqlx.DB {
	return b.db
}

func (b Backends) KYC() kyc.Client {
	return b.kyc
}

func (b Backends) LinkedAccounts() linkedaccounts.Client {
	return b.la
}

func (b Backends) Users() user.Client {
	return b.users
}

func (b Backends) Temporal() temporal.Client {
	return b.tp
}

func (b Backends) Pacioli() pacioli.Client {
	return nil
}
func (b Backends) Transactions() transactions.Client {
	return b.txc
}
func (b Backends) Wallets() wallets.Client {
	return b.wc
}

func NewBackends(t *testing.T) *Backends {
	ctrl := gomock.NewController(t)
	t.Cleanup(func() {
		ctrl.Finish()
	})

	return &Backends{
		db:    db.MigrateTestDB(t, context.Background()),
		kyc:   kyc_mock.NewMockClient(ctrl),
		la:    la_mock.NewMockClient(ctrl),
		users: user_mock.NewMock(),
		pc:    payments_mock.NewMockClient(ctrl),
	}
}
