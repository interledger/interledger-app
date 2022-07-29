package ops

import (
	"github.com/go-playground/validator/v10"
	"github.com/jmoiron/sqlx"
	"gitlab.com/fynbos/backend/accounts"
	"gitlab.com/fynbos/pacioli"
)

type Backends interface {
	Validator() *validator.Validate
	DB() *sqlx.DB
	Pacioli() pacioli.Client
	Accounts() accounts.Client
}
