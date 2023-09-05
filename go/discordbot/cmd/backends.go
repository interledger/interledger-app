package cmd

import (
	"gitlab.com/fynbos/backend/linkedaccounts"
	"gitlab.com/fynbos/backend/payments"

	"gitlab.com/fynbos/backend/identities"
	"gitlab.com/fynbos/backend/wallets"
)

type Backends interface {
	LinkedAccounts() linkedaccounts.Client
	Wallets() wallets.Client
	Identities() identities.Client
	Payments() payments.Client
}
