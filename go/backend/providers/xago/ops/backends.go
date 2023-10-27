package ops

import (
	"github.com/jmoiron/sqlx"
	"gitlab.com/fynbos/backend/linkedaccounts"
	"gitlab.com/fynbos/backend/payments"
	"gitlab.com/fynbos/backend/providers/xago/external"
	"gitlab.com/fynbos/backend/wallets"
)

type Backends interface {
	DB() *sqlx.DB
	External() external.Client
	Payments() payments.Client
	LinkedAccounts() linkedaccounts.Client
	Wallets() wallets.Client
}
