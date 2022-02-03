//go:generate go run github.com/99designs/gqlgen

package graph

// This file will not be regenerated automatically.
//
// It serves as dependency injection for your app, add any dependencies you require here.

import (
	"github.com/jmoiron/sqlx"
	"gitlab.com/fynbos/backend/accounts"
	"gitlab.com/fynbos/backend/identity"
	"gitlab.com/fynbos/backend/user"
)

type Resolver struct {
	// appending service to avoid name clashed with function names.
	IdentityService identity.Service
	UserService     user.Service
	AccountService  accounts.Service
	Db              *sqlx.DB
}
