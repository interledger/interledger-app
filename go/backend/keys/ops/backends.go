package ops

import (
	"testing"

	"gitlab.com/fynbos/backend/rafiki"
	"gitlab.com/fynbos/backend/vault"

	"github.com/go-playground/validator/v10"
	"github.com/jmoiron/sqlx"
)

type Backends interface {
	Validator() *validator.Validate
	DB() *sqlx.DB
	Vault() vault.Client
	Rafiki() rafiki.Client
}

type testBackends struct {
	db     *sqlx.DB
	val    *validator.Validate
	vault  vault.Client
	rafiki rafiki.Client
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

func (t testBackends) Rafiki() rafiki.Client {
	return t.rafiki
}

func NewTestBackends(_ *testing.T, db *sqlx.DB, vc vault.Client, rc rafiki.Client) Backends {
	return &testBackends{db: db, val: validator.New(), vault: vc, rafiki: rc}
}
