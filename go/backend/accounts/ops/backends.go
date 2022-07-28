package ops

import (
	"testing"

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

func NewTestBackends(_ *testing.T, db *sqlx.DB,
	ids identity.Service,
	countries country.Service,
	pacioli pacioli.Client) Backends {

	return &backends{
		val:       validator.New(),
		db:        db,
		ids:       ids,
		countries: countries,
		pacioli:   pacioli,
	}
}

var _ Backends = backends{}

type backends struct {
	val       *validator.Validate
	db        *sqlx.DB
	ids       identity.Service
	countries country.Service
	pacioli   pacioli.Client
}

func (b backends) Identity() identity.Service {
	return b.ids
}

func (b backends) Countries() country.Service {
	return b.countries
}

func (b backends) Pacioli() pacioli.Client {
	return b.pacioli
}

func (b backends) DB() *sqlx.DB {
	return b.db
}

func (b backends) Validator() *validator.Validate {
	return b.val
}
