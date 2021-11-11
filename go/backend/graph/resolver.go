//go:generate go run github.com/99designs/gqlgen

package graph

// This file will not be regenerated automatically.
//
// It serves as dependency injection for your app, add any dependencies you require here.

import (
	org "gitlab.com/fynbos/backend/organisation"
	"gitlab.com/fynbos/backend/user"
)

type Resolver struct {
	Organisations org.Service
	User          user.Service
}
