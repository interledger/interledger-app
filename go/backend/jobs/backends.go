package jobs

import (
	"github.com/interledger/interledger-app/go/backend/email"
	"github.com/interledger/interledger-app/go/backend/keys"
	"github.com/interledger/interledger-app/go/backend/kyc"
	"github.com/interledger/interledger-app/go/backend/linkedaccounts"
	"github.com/interledger/interledger-app/go/backend/payments"
	"github.com/interledger/interledger-app/go/backend/providers/gatehub"
	"github.com/interledger/interledger-app/go/backend/providers/pti"
	"github.com/interledger/interledger-app/go/backend/rafiki"
	"github.com/interledger/interledger-app/go/backend/transactions"
	"github.com/interledger/interledger-app/go/backend/user"
	"github.com/interledger/interledger-app/go/backend/wallets"
	"github.com/interledger/interledger-app/go/pacioli"
	"github.com/jmoiron/sqlx"
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
	PTI() pti.Client
}

type Config struct {
	KratosURL         string
	KratosAdminURL    string
	PTIJWK            string
	PTIBaseURL        string
	PTIClientID       string
	RafikiDBURL       string
	RafikiAuthDBURL   string
	TempGatehubAppID  string
	TempGatehubSecret string
}

type Activity struct {
	b             Backends
	gatehubConfig gatehub.Config
	cfg           Config
}

func NewActivity(b Backends, gatehubConfig gatehub.Config, cfg Config) *Activity {
	return &Activity{
		b:             b,
		gatehubConfig: gatehubConfig,
		cfg:           cfg,
	}
}
