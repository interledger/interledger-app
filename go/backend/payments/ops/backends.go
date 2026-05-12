package ops

import (
	"testing"

	"gitlab.com/fynbos/backend/providers/chimoney"

	"gitlab.com/fynbos/backend/providers/gatehub"
	"gitlab.com/fynbos/backend/providers/pti"
	pti_mock "gitlab.com/fynbos/backend/providers/pti/client/mock"
	"gitlab.com/fynbos/pacioli"

	"gitlab.com/fynbos/backend/limits"

	"gitlab.com/fynbos/backend/providers/xago"

	"gitlab.com/fynbos/backend/rafiki"

	"github.com/go-playground/validator/v10"
	"github.com/jmoiron/sqlx"
	"gitlab.com/fynbos/backend/email"
	email_mock "gitlab.com/fynbos/backend/email/client/mock"
	"gitlab.com/fynbos/backend/identities"
	id_mock "gitlab.com/fynbos/backend/identities/client/mock"
	"gitlab.com/fynbos/backend/linkedaccounts"
	linkedaccounts_mock "gitlab.com/fynbos/backend/linkedaccounts/client/mock"
	"gitlab.com/fynbos/backend/notify"
	temporal_mock "gitlab.com/fynbos/backend/temporal/mock"
	"gitlab.com/fynbos/backend/transactions"
	transactions_mock "gitlab.com/fynbos/backend/transactions/client/mock"
	"gitlab.com/fynbos/backend/user"
	"gitlab.com/fynbos/backend/wallets"
	wallets_mock "gitlab.com/fynbos/backend/wallets/client/mock"
	temporal "go.temporal.io/sdk/client"
)

type Backends interface {
	Validator() *validator.Validate
	DB() *sqlx.DB
	Temporal() temporal.Client
	Notify() notify.Client
	Email() email.Client
	Wallets() wallets.Client
	Identities() identities.Client
	LinkedAccounts() linkedaccounts.Client
	Transactions() transactions.Client
	Rafiki() rafiki.Client
	Xago() xago.Client
	Limits() limits.Client
	PTI() pti.Client
	Gatehub() gatehub.Client
	Chimoney() chimoney.Client
	Pacioli() pacioli.Client
}

type TestBackends struct {
	DBC *sqlx.DB
	val *validator.Validate
	Tp  *temporal_mock.MockClient
	uc  user.Client
	nc  notify.Client
	Em  *email_mock.MockClient
	Wc  *wallets_mock.MockClient
	Ic  *id_mock.MockClient
	Lac *linkedaccounts_mock.MockClient
	Txc *transactions_mock.MockClient
	Pti *pti_mock.MockClient
}

func (t TestBackends) Chimoney() chimoney.Client {
	return nil
}

func (t TestBackends) Gatehub() gatehub.Client {
	return nil
}

func (t TestBackends) PTI() pti.Client {
	return t.Pti
}

func (t TestBackends) Limits() limits.Client {
	return nil
}

func (t TestBackends) Xago() xago.Client {
	return nil
}

func (t TestBackends) Rafiki() rafiki.Client {
	return nil
}

func (t TestBackends) Transactions() transactions.Client {
	return t.Txc
}

func (t TestBackends) LinkedAccounts() linkedaccounts.Client {
	return t.Lac
}

func (t TestBackends) Identities() identities.Client {
	return t.Ic
}

func (t TestBackends) Users() user.Client {
	return t.uc
}

func (t TestBackends) Validator() *validator.Validate {
	if t.val == nil {
		t.val = validator.New()
	}
	return t.val
}

func (t TestBackends) DB() *sqlx.DB {
	return t.DBC
}

func (t TestBackends) Temporal() temporal.Client {
	return t.Tp
}

func (t TestBackends) Notify() notify.Client {
	return t.nc
}

func (t TestBackends) Email() email.Client {
	return t.Em
}

func (t TestBackends) Wallets() wallets.Client {
	return t.Wc
}

func (t TestBackends) Pacioli() pacioli.Client {
	return nil
}

func NewTestBackends(_ *testing.T, opts ...func(*TestBackends)) Backends {
	b := &TestBackends{val: validator.New()}
	for _, opt := range opts {
		opt(b)
	}
	return b
}
