package webhook

import (
	"gitlab.com/fynbos/backend/analytics"
	"gitlab.com/fynbos/backend/user"
	"gitlab.com/fynbos/backend/wallets"
)

type Backends interface {
	Users() user.Client
	Wallets() wallets.Client
	Analytics() analytics.Client
}
