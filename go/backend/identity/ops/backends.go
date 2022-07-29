package ops

import (
	"github.com/go-playground/validator/v10"
	"github.com/jmoiron/sqlx"
	"gitlab.com/fynbos/backend/country"
)

type Backends interface {
	Validator() *validator.Validate
	DB() *sqlx.DB
	Countries() country.Client
}
