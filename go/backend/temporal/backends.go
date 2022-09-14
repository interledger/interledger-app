package temporal

import (
	"github.com/go-playground/validator/v10"
	"github.com/jmoiron/sqlx"
	"gitlab.com/fynbos/backend/accounts"
	transactions "gitlab.com/fynbos/backend/accounttransactions"
	"gitlab.com/fynbos/backend/deposits"
	"gitlab.com/fynbos/backend/fundingsources"
	"gitlab.com/fynbos/backend/identity"
	"gitlab.com/fynbos/backend/payments"
	"gitlab.com/fynbos/backend/providers/mx"
	"gitlab.com/fynbos/backend/providers/noop"
	"gitlab.com/fynbos/backend/providers/unit"
	"gitlab.com/fynbos/backend/twilio"
	"gitlab.com/fynbos/backend/withdrawals"
	"go.temporal.io/sdk/client"
)

type Backends interface {
	Validator() *validator.Validate
	DB() *sqlx.DB
	Accounts() accounts.Client
	Deposits() deposits.Service
	Identity() identity.Client
	Temporal() client.Client
	Payments() payments.Client
	Transactions() transactions.Client
	Noop() noop.Service
	Twilio() twilio.Service
	MX() mx.Client
	Unit() unit.Client
	FundingSources() fundingsources.Client
	Withdrawals() withdrawals.Service
}
