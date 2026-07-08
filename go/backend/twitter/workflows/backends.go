package workflows

import (
	"github.com/interledger/interledger-app/go/backend/identities"
	"github.com/interledger/interledger-app/go/backend/twitter"
	"github.com/interledger/interledger-app/go/backend/wallets"
)

type Backends interface {
	Twitter() twitter.Client
	Identities() identities.Client
	Wallets() wallets.Client
}
