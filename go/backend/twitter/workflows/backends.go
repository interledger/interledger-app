package workflows

import (
	"gitlab.com/fynbos/backend/identities"
	"gitlab.com/fynbos/backend/twitter"
	"gitlab.com/fynbos/backend/wallets"
)

type Backends interface {
	Twitter() twitter.Client
	Identities() identities.Client
	Wallets() wallets.Client
}
