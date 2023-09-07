package bot

import (
	"github.com/jmoiron/sqlx"
	"gitlab.com/fynbos/backend/linkedaccounts"
	"gitlab.com/fynbos/backend/payments"
	"gitlab.com/fynbos/backend/wallets"
)

type Backends interface {
	DB() *sqlx.DB
	Payments() payments.Client
	Wallets() wallets.Client
	LinkedAccounts() linkedaccounts.Client
}
