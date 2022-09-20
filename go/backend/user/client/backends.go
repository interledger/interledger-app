package client

import (
	"testing"

	"github.com/go-playground/validator/v10"
	"github.com/jmoiron/sqlx"
	kratos "github.com/ory/kratos-client-go"
	"gitlab.com/fynbos/backend/user/ops"
)

type Backends interface {
	Validator() *validator.Validate
	DB() *sqlx.DB
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
) ops.Backends {

	return &testBackends{
		val: validator.New(),
		db:  db,
		kr:  kratos,
	}
}

var _ ops.Backends = testBackends{}

type testBackends struct {
	val *validator.Validate
	db  *sqlx.DB
	kr  *kratos.APIClient
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
