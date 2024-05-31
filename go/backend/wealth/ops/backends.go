package ops

import (
	"github.com/jmoiron/sqlx"
	"gitlab.com/fynbos/backend/vault"
	temporal "go.temporal.io/sdk/client"
)

type Backends interface {
	DB() *sqlx.DB
	Vault() vault.Client
	Temporal() temporal.Client
}
