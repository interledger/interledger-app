package jobs

import (
	"github.com/jmoiron/sqlx"
	"gitlab.com/fynbos/backend/email"
	"gitlab.com/fynbos/backend/keys"
	"gitlab.com/fynbos/backend/kyc"
	"gitlab.com/fynbos/backend/linkedaccounts"
	"gitlab.com/fynbos/backend/payments"
	"gitlab.com/fynbos/backend/providers/gatehub"
	"gitlab.com/fynbos/backend/rafiki"
	"gitlab.com/fynbos/backend/transactions"
	"gitlab.com/fynbos/backend/user"
	"gitlab.com/fynbos/backend/wallets"
	"gitlab.com/fynbos/pacioli"
)

type Backends interface {
	DB() *sqlx.DB
	Users() user.Client
	Keys() keys.Client
	KYC() kyc.Client
	Gatehub() gatehub.Client
	Wallets() wallets.Client
	Transactions() transactions.Client
	Rafiki() rafiki.Client
	Email() email.Client
	Payments() payments.Client
	LinkedAccounts() linkedaccounts.Client
	Pacioli() pacioli.Client
}

type Activity struct {
	b Backends
}

func NewActivity(b Backends) *Activity {
	return &Activity{b: b}
}
