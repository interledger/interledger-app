//go:generate go run github.com/99designs/gqlgen

package graph

// This file will not be regenerated automatically.
//
// It serves as dependency injection for your app, add any dependencies you require here.

import (
	"github.com/jmoiron/sqlx"
	"gitlab.com/fynbos/backend/accounts"
	account_transactions "gitlab.com/fynbos/backend/accounttransactions"
	"gitlab.com/fynbos/backend/country"
	"gitlab.com/fynbos/backend/deposits"
	"gitlab.com/fynbos/backend/fundingsources"
	"gitlab.com/fynbos/backend/identity"
	"gitlab.com/fynbos/backend/onboarding"
	"gitlab.com/fynbos/backend/providers/noop"
	"gitlab.com/fynbos/backend/providers/unit"
	"gitlab.com/fynbos/backend/user"
)

type Resolver struct {
	// appending service to avoid name clashed with function names.
	IdentityService     identity.Client
	UserService         user.Client
	CountryService      country.Client
	AccountService      accounts.Client
	NoopService         noop.Service
	UnitService         unit.Client
	Db                  *sqlx.DB
	AccountTransactions account_transactions.Client
	Ds                  deposits.Service
	Os                  onboarding.Client
	Fs                  fundingsources.Client
}
