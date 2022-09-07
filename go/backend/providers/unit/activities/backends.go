package activities

import (
	"github.com/go-playground/validator/v10"
	"gitlab.com/fynbos/backend/accounts"
	"gitlab.com/fynbos/backend/identity"
	"gitlab.com/fynbos/backend/providers/unit"
)

type Backends interface {
	Validator() *validator.Validate
	Unit() unit.Client
	Accounts() accounts.Client
	Identity() identity.Client
}
