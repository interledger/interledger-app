package temporal

import (
	"github.com/go-playground/validator/v10"
	"github.com/jmoiron/sqlx"
	"gitlab.com/fynbos/backend/accounts"
	transactions "gitlab.com/fynbos/backend/accounttransactions"
	"gitlab.com/fynbos/backend/fundingsources"
	"gitlab.com/fynbos/backend/identity"
	"gitlab.com/fynbos/backend/providers/mx"
	"gitlab.com/fynbos/backend/twilio"
	"go.temporal.io/sdk/client"
)

type Backends interface {
	Validator() *validator.Validate
	DB() *sqlx.DB
	Accounts() accounts.Client
	Identity() identity.Client
	Temporal() client.Client
	Transactions() transactions.Client
	Twilio() twilio.Service
	MX() mx.Client
	FundingSources() fundingsources.Client
}
