package ops

import (
	"testing"

	"github.com/jmoiron/sqlx"
	"gitlab.com/fynbos/backend/kyc"
)

type Backends interface {
	DB() *sqlx.DB
	KYC() kyc.Client
}

type testBackends struct {
	db  *sqlx.DB
	kyc kyc.Client
}

func (t testBackends) DB() *sqlx.DB {
	return t.db
}

func (t testBackends) KYC() kyc.Client {
	return t.kyc
}

func NewTestBackends(t *testing.T, db *sqlx.DB, kyc kyc.Client) Backends {
	return &testBackends{db: db, kyc: kyc}
}
