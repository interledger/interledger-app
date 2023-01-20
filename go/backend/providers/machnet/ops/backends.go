package ops

import (
	"github.com/jmoiron/sqlx"
	"gitlab.com/fynbos/backend/email"
	"gitlab.com/fynbos/backend/kyc"
	"gitlab.com/fynbos/backend/linkedaccounts"
	"gitlab.com/fynbos/backend/providers/machnet/external"
	"gitlab.com/fynbos/backend/statements"
	"gitlab.com/fynbos/backend/transactions"
	"gitlab.com/fynbos/backend/user"
	"go.temporal.io/sdk/client"
)

type Backends interface {
	DB() *sqlx.DB
	Users() user.Client
	KYC() kyc.Client
	External() external.Client
	LinkedAccounts() linkedaccounts.Client
	Email() email.Client
	Statements() statements.Client
	Temporal() client.Client
	Transactions() transactions.Client
}
