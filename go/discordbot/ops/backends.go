package ops

import (
	"github.com/jmoiron/sqlx"
	"gitlab.com/fynbos/backend/identities"
	"gitlab.com/fynbos/backend/linkedaccounts"
	"gitlab.com/fynbos/backend/payments"
	"gitlab.com/fynbos/backend/transactions"
	"gitlab.com/fynbos/backend/wallets"
)

type Backends interface {
	DB() *sqlx.DB
	Discord() Discord
	Payments() payments.Client
	LinkedAccounts() linkedaccounts.Client
	Identities() identities.Client
	Transactions() transactions.Client
	Wallets() wallets.Client
}
