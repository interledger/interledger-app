package ops

import (
	"gitlab.com/fynbos/backend/vault"
	"testing"

	"github.com/go-playground/validator/v10"
	"github.com/jmoiron/sqlx"
)

type Backends interface {
	Validator() *validator.Validate
	DB() *sqlx.DB
	Vault() vault.Client
}

type testBackends struct {
	db    *sqlx.DB
	val   *validator.Validate
	vault vault.Client
}

func (t testBackends) Validator() *validator.Validate {
	return t.val
}

func (t testBackends) DB() *sqlx.DB {
	return t.db
}

func (t testBackends) Vault() vault.Client {
	return t.vault
}

func NewTestBackends(_ *testing.T, db *sqlx.DB, vc vault.Client) Backends {
	return &testBackends{db: db, val: validator.New(), vault: vc}
}
