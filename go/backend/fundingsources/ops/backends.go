package ops

import (
	"github.com/go-playground/validator/v10"
	"github.com/jmoiron/sqlx"
	"gitlab.com/fynbos/backend/accounts"
	"gitlab.com/fynbos/backend/identity"
	"gitlab.com/fynbos/backend/providers/noop"
	"gitlab.com/fynbos/backend/providers/unit"
	"gitlab.com/fynbos/pacioli"
	temporal "go.temporal.io/sdk/client"
)

type Backends interface {
	Validator() *validator.Validate
	DB() *sqlx.DB
	Pacioli() pacioli.Client
	Accounts() accounts.Client
	Identity() identity.Service
	Noop() noop.Service
	Temporal() temporal.Client
	Unit() unit.Service
}
