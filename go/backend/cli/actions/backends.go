package actions

import (
	"github.com/go-playground/validator/v10"
	"github.com/jmoiron/sqlx"
	kratos "github.com/ory/kratos-client-go"
	"gitlab.com/fynbos/backend/kyc"
	"gitlab.com/fynbos/backend/linkedaccounts"
	"gitlab.com/fynbos/backend/providers/machnet"
	machnet_external "gitlab.com/fynbos/backend/providers/machnet/external"
	"gitlab.com/fynbos/backend/signup"
	"gitlab.com/fynbos/backend/user"
)

type Backends interface {
	DB() *sqlx.DB
	Kratos() *kratos.APIClient
	KYC() kyc.Client
	LinkedAccounts() linkedaccounts.Client
	Machnet() machnet.Client
	MachnetExternal() machnet_external.Client
	Signup() signup.Client
	Users() user.Client
	Validator() *validator.Validate
}
