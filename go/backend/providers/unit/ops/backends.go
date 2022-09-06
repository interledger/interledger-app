package ops

import (
	"testing"

	"github.com/go-playground/validator/v10"
	"github.com/jmoiron/sqlx"
	"gitlab.com/fynbos/backend/accounts"
	"gitlab.com/fynbos/backend/identity"
	"go.temporal.io/sdk/client"
)

type Backends interface {
	Validator() *validator.Validate
	DB() *sqlx.DB
	Identity() identity.Client
	Accounts() accounts.Client
	Temporal() client.Client
}

func NewTestBackends(
	_ *testing.T,
	db *sqlx.DB,
	ids identity.Client,
	accounts accounts.Client,
	temporal client.Client,
) Backends {

	return &backends{
		val:      validator.New(),
		db:       db,
		ids:      ids,
		accounts: accounts,
		temporal: temporal,
	}
}

var _ backends = backends{}

type backends struct {
	val      *validator.Validate
	db       *sqlx.DB
	ids      identity.Client
	accounts accounts.Client
	temporal client.Client
}

func (b backends) Validator() *validator.Validate {
	return b.val
}

func (b backends) DB() *sqlx.DB {
	return b.db
}

func (b backends) Identity() identity.Client {
	return b.ids
}

func (b backends) Accounts() accounts.Client {
	return b.accounts
}

func (b backends) Temporal() client.Client {
	return b.temporal
}
