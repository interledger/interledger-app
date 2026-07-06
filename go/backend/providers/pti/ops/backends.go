package ops

import (
	"github.com/interledger/interledger-app/go/backend/kyc"
	"github.com/interledger/interledger-app/go/backend/linkedaccounts"
	"github.com/interledger/interledger-app/go/backend/payments"
	"github.com/interledger/interledger-app/go/backend/transactions"
	"github.com/interledger/interledger-app/go/backend/user"
	"github.com/interledger/interledger-app/go/backend/wallets"
	"github.com/interledger/interledger-app/go/pacioli"
	"github.com/jmoiron/sqlx"
	temporal "go.temporal.io/sdk/client"
)

type Backends interface {
	DB() *sqlx.DB
	KYC() kyc.Client
	LinkedAccounts() linkedaccounts.Client
	Users() user.Client
	Temporal() temporal.Client
	Payments() payments.Client
	Pacioli() pacioli.Client
	Transactions() transactions.Client
	Wallets() wallets.Client
}
