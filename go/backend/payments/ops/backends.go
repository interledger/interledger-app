package ops

import (
	"testing"

	wallets_mock "gitlab.com/fynbos/backend/wallets/client/mock"

	"gitlab.com/fynbos/backend/transactions"

	"gitlab.com/fynbos/backend/providers/tabapay"

	"gitlab.com/fynbos/backend/linkedaccounts"

	"gitlab.com/fynbos/backend/identities"
	temporal "go.temporal.io/sdk/client"

	"gitlab.com/fynbos/backend/email"
	email_mock "gitlab.com/fynbos/backend/email/client/mock"
	id_mock "gitlab.com/fynbos/backend/identities/client/mock"
	"gitlab.com/fynbos/backend/notify"
	"gitlab.com/fynbos/backend/wallets"

	"gitlab.com/fynbos/backend/user"

	temporal_mock "gitlab.com/fynbos/backend/temporal/mock"

	"github.com/go-playground/validator/v10"
	"github.com/jmoiron/sqlx"
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
	Tabapay() tabapay.Client
	Transactions() transactions.Client
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
}

func (t TestBackends) Transactions() transactions.Client {
	//TODO implement me
	panic("implement me")
}

func (t TestBackends) Tabapay() tabapay.Client {
	//TODO implement me
	panic("implement me")
}

func (t TestBackends) LinkedAccounts() linkedaccounts.Client {
	//TODO implement me
	panic("implement me")
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

func NewTestBackends(_ *testing.T, opts ...func(*TestBackends)) Backends {
	b := &TestBackends{val: validator.New()}
	for _, opt := range opts {
		opt(b)
	}
	return b
}
