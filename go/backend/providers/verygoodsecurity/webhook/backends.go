package webhook

import (
	"github.com/jmoiron/sqlx"
	"gitlab.com/fynbos/backend/analytics"
	"gitlab.com/fynbos/backend/email"
	"gitlab.com/fynbos/backend/kyc"
	"gitlab.com/fynbos/backend/linkedaccounts"
	"gitlab.com/fynbos/backend/notify"
	"gitlab.com/fynbos/backend/providers/verygoodsecurity"
	"gitlab.com/fynbos/backend/statements"
	"gitlab.com/fynbos/backend/transactions"
	"gitlab.com/fynbos/backend/user"
	"go.temporal.io/sdk/client"
)

type Backends interface {
	DB() *sqlx.DB
	Users() user.Client
	KYC() kyc.Client
	VGS() verygoodsecurity.Client
	LinkedAccounts() linkedaccounts.Client
	Statements() statements.Client
	Temporal() client.Client
	Transactions() transactions.Client
	Email() email.Client
	Analytics() analytics.Client
	Notify() notify.Client
}
