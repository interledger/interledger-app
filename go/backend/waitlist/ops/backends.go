package ops

import (
	"testing"

	"github.com/go-playground/validator/v10"
	"github.com/jmoiron/sqlx"
)

type Backends interface {
	DB() *sqlx.DB
	Validator() *validator.Validate
}

type backends struct {
	db  *sqlx.DB
	val *validator.Validate
}

func (b backends) DB() *sqlx.DB {
	return b.db
}

func (b backends) Validator() *validator.Validate {
	return b.val
}

func NewBackends(_ *testing.T, db *sqlx.DB, val *validator.Validate) Backends {
	return &backends{
		db:  db,
		val: val,
	}
}
