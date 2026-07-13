package ops

import (
	"testing"

	"github.com/interledger/interledger-app/go/backend/config"
	"github.com/interledger/interledger-app/go/backend/rafiki"
	"github.com/interledger/interledger-app/go/backend/vault"

	"github.com/go-playground/validator/v10"
	"github.com/jmoiron/sqlx"
)

type Backends interface {
	Validator() *validator.Validate
	DB() *sqlx.DB
	Vault() vault.Client
	Rafiki() rafiki.Client
	Config() *config.StartConfig
}

type testBackends struct {
	db     *sqlx.DB
	val    *validator.Validate
	vault  vault.Client
	rafiki rafiki.Client
	cfg    *config.StartConfig
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

func (t testBackends) Config() *config.StartConfig {
	return t.cfg
}

func NewTestBackends(_ *testing.T, db *sqlx.DB, vc vault.Client, rc rafiki.Client, cfg *config.StartConfig) Backends {
	return &testBackends{db: db, val: validator.New(), vault: vc, rafiki: rc, cfg: cfg}
}
