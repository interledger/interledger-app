package ops_test

import (
	"github.com/interledger/interledger-app/go/backend/email"
	"github.com/interledger/interledger-app/go/backend/kyc"
	"github.com/interledger/interledger-app/go/backend/linkedaccounts"
	"github.com/interledger/interledger-app/go/backend/providers/chimoney/ops"
	"github.com/interledger/interledger-app/go/backend/transactions"
	"github.com/interledger/interledger-app/go/backend/user"
	"github.com/interledger/interledger-app/go/backend/wallets"
	"github.com/interledger/interledger-app/go/pacioli"
	"github.com/jmoiron/sqlx"
	temporal "go.temporal.io/sdk/client"
)

var _ ops.Backends = backends{}

type backends struct {
	db           *sqlx.DB
	las          linkedaccounts.Client
	users        user.Client
	temporal     temporal.Client
	wallets      wallets.Client
	pacioli      pacioli.Client
	kyc          kyc.Client
	transactions transactions.Client
	email        email.Client
}

func (b backends) Email() email.Client {
	return b.email
}

func (b backends) DB() *sqlx.DB {
	return b.db
}

func (b backends) LinkedAccounts() linkedaccounts.Client {
	return b.las
}

func (b backends) Users() user.Client {
	return b.users
}

func (b backends) Temporal() temporal.Client {
	return b.temporal
}

func (b backends) Wallets() wallets.Client {
	return b.wallets
}

func (b backends) Pacioli() pacioli.Client {
	return b.pacioli
}

func (b backends) KYC() kyc.Client {
	return b.kyc
}

func (b backends) Transactions() transactions.Client {
	return b.transactions
}
