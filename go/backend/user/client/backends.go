package client

import (
	"github.com/interledger/interledger-app/go/backend/analytics"
	analytics_client "github.com/interledger/interledger-app/go/backend/analytics/client"
	"github.com/interledger/interledger-app/go/backend/keys"
	"testing"

	"github.com/go-playground/validator/v10"
	"github.com/interledger/interledger-app/go/backend/user/ops"
	"github.com/jmoiron/sqlx"
	kratos "github.com/ory/kratos-client-go"
)

type Backends interface {
	Validator() *validator.Validate
	DB() *sqlx.DB
	Analytics() analytics.Client
	Keys() keys.Client
}

var _ ops.Backends = opsBackends{}

type opsBackends struct {
	Backends
	kratos *kratos.APIClient
}

func (ob opsBackends) Kratos() *kratos.APIClient {
	return ob.kratos
}

func NewTestBackends(
	_ *testing.T,
	db *sqlx.DB,
	kratos *kratos.APIClient,
	keys keys.Client,
) ops.Backends {
	return &testBackends{
		val: validator.New(),
		db:  db,
		kr:  kratos,
		ac:  analytics_client.New(nil, ""),
		kc:  keys,
	}
}

var _ ops.Backends = testBackends{}

type testBackends struct {
	val *validator.Validate
	db  *sqlx.DB
	kr  *kratos.APIClient
	ac  analytics.Client
	kc  keys.Client
}

func (b testBackends) Kratos() *kratos.APIClient {
	return b.kr
}

func (b testBackends) Validator() *validator.Validate {
	return b.val
}

func (b testBackends) DB() *sqlx.DB {
	return b.db
}

func (b testBackends) Analytics() analytics.Client {
	return b.ac
}

func (b testBackends) Keys() keys.Client {
	return b.kc
}
