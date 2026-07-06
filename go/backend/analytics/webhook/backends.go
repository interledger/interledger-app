package webhook

import (
	"github.com/interledger/interledger-app/go/backend/analytics"
	"github.com/interledger/interledger-app/go/backend/user"
	"github.com/interledger/interledger-app/go/backend/wallets"
)

type Backends interface {
	Users() user.Client
	Wallets() wallets.Client
	Analytics() analytics.Client
}
