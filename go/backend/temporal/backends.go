package temporal

import (
	"github.com/go-playground/validator/v10"
	"github.com/jmoiron/sqlx"
	"gitlab.com/fynbos/backend/accounts"
	transactions "gitlab.com/fynbos/backend/accounttransactions"
	"gitlab.com/fynbos/backend/identity"
	"gitlab.com/fynbos/backend/payments"
	"gitlab.com/fynbos/backend/providers/noop"
	"gitlab.com/fynbos/backend/providers/unit"
	"go.temporal.io/sdk/client"
)

type Backends interface {
	Validator() *validator.Validate
	DB() *sqlx.DB
	Accounts() accounts.Client
	Identity() identity.Client
	Temporal() client.Client
	Payments() payments.Client
	Transactions() transactions.Client
	Noop() noop.Service
	Unit() unit.Client
}
