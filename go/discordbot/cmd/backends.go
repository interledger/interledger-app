package cmd

import (
	"github.com/jmoiron/sqlx"
	"gitlab.com/fynbos/backend/linkedaccounts"
	"gitlab.com/fynbos/backend/payments"
	"gitlab.com/fynbos/discordbot/ops"

	"gitlab.com/fynbos/backend/identities"
	"gitlab.com/fynbos/backend/wallets"
)

type Backends interface {
	DB() *sqlx.DB
	LinkedAccounts() linkedaccounts.Client
	Wallets() wallets.Client
	Identities() identities.Client
	Payments() payments.Client
	Discord() ops.Discord
}
