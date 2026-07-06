package ops

import (
	"testing"

	"github.com/go-playground/validator/v10"
	"github.com/interledger/interledger-app/go/backend/analytics"
	analytics_client "github.com/interledger/interledger-app/go/backend/analytics/client"
	"github.com/interledger/interledger-app/go/backend/keys"
	"github.com/interledger/interledger-app/go/backend/user"
	"github.com/jmoiron/sqlx"
)

type Backends interface {
	Validator() *validator.Validate
	DB() *sqlx.DB
	Analytics() analytics.Client
	Keys() keys.Client
	Users() user.Client
}

type testBackends struct {
	val *validator.Validate
	db  *sqlx.DB
	ac  analytics.Client
	kc  keys.Client
	uc  user.Client
}

func NewTestBackends(
	_ *testing.T,
	db *sqlx.DB,
	keys keys.Client,
	uc user.Client,
) Backends {
	return &testBackends{
		val: validator.New(),
		db:  db,
		ac:  analytics_client.New(nil, ""),
		kc:  keys,
		uc:  uc,
	}
}

var _ Backends = testBackends{}

func (t testBackends) Validator() *validator.Validate {
	return t.val
}

func (t testBackends) DB() *sqlx.DB {
	return t.db
}

func (t testBackends) Analytics() analytics.Client {
	return t.ac
}

func (t testBackends) Keys() keys.Client {
	return t.kc
}

func (t testBackends) Users() user.Client {
	return t.uc
}
