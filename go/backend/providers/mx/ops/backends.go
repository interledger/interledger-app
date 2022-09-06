package ops

import (
	"github.com/go-playground/validator/v10"
	"github.com/jmoiron/sqlx"
	"gitlab.com/fynbos/backend/accounts"
	"gitlab.com/fynbos/backend/identity"
	"gitlab.com/fynbos/backend/providers/mx/external"
	"gitlab.com/fynbos/backend/twilio"
	"go.temporal.io/sdk/client"
)

type Backends interface {
	Validator() *validator.Validate
	DB() *sqlx.DB
	Accounts() accounts.Client
	Identity() identity.Client
	Temporal() client.Client
	Twilio() twilio.Service
	MXExternal() external.Mx
}
