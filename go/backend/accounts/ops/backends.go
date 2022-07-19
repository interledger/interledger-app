package ops

import (
	"github.com/go-playground/validator/v10"
	"github.com/jmoiron/sqlx"
	"gitlab.com/fynbos/backend/country"
	"gitlab.com/fynbos/backend/identity"
	"gitlab.com/fynbos/pacioli"
)

type Backends interface {
	Validator() *validator.Validate
	DB() *sqlx.DB
	Identity() identity.Service
	Countries() country.Service
	Pacioli() pacioli.Client
}
