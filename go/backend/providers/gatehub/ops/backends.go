package ops

import (
	"github.com/jmoiron/sqlx"
	"gitlab.com/fynbos/backend/email"
	"gitlab.com/fynbos/backend/kyc"
	"gitlab.com/fynbos/backend/linkedaccounts"
	"gitlab.com/fynbos/backend/payments"
	"gitlab.com/fynbos/backend/transactions"
	"gitlab.com/fynbos/backend/user"
	"gitlab.com/fynbos/backend/wallets"
	"gitlab.com/fynbos/pacioli"
	temporal "go.temporal.io/sdk/client"
)

type Backends interface {
	DB() *sqlx.DB
	LinkedAccounts() linkedaccounts.Client
	Users() user.Client
	Temporal() temporal.Client
	Wallets() wallets.Client
	Pacioli() pacioli.Client
	Email() email.Client
	KYC() kyc.Client
	Transactions() transactions.Client
	Payments() payments.Client
}
