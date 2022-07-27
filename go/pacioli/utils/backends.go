package test_utils

import (
	"testing"

	tigerbeetle_go "github.com/coilhq/tigerbeetle-go"
	"github.com/go-playground/validator/v10"
	"github.com/jmoiron/sqlx"
)

type Backends interface {
	DB() *sqlx.DB
	TigerBeetle() tigerbeetle_go.Client
	Validator() *validator.Validate
}

var _ Backends = backends{}

type backends struct {
	db  *sqlx.DB
	tbc tigerbeetle_go.Client
	val *validator.Validate
}

func NewBackends(_ *testing.T, db *sqlx.DB, tbc tigerbeetle_go.Client) Backends {
	return &backends{
		db:  db,
		tbc: tbc,
		val: validator.New(),
	}
}

func (b backends) DB() *sqlx.DB {
	return b.db
}

func (b backends) TigerBeetle() tigerbeetle_go.Client {
	return b.tbc
}

func (b backends) Validator() *validator.Validate {
	return b.val
}
