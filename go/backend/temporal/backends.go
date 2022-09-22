package temporal

import (
	"github.com/go-playground/validator/v10"
	"github.com/jmoiron/sqlx"
	"gitlab.com/fynbos/backend/linkedaccounts"
	"gitlab.com/fynbos/backend/twilio"
	"go.temporal.io/sdk/client"
)

type Backends interface {
	Validator() *validator.Validate
	DB() *sqlx.DB
	Temporal() client.Client
	Twilio() twilio.Service
	LinkedAccounts() linkedaccounts.Client
}
