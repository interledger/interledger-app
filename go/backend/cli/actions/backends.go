package actions

import (
	"github.com/go-playground/validator/v10"
	"github.com/jmoiron/sqlx"
	kratos "github.com/ory/kratos-client-go"
	"gitlab.com/fynbos/backend/kyc"
	"gitlab.com/fynbos/backend/linkedaccounts"
	"gitlab.com/fynbos/backend/user"
	"gitlab.com/fynbos/backend/wallets"
)

type Backends interface {
	DB() *sqlx.DB
	Kratos() *kratos.APIClient
	KYC() kyc.Client
	LinkedAccounts() linkedaccounts.Client
	Users() user.Client
	Validator() *validator.Validate
	Wallets() wallets.Client
}
