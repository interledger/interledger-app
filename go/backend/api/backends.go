package api

import (
	"gitlab.com/fynbos/backend/providers/gatehub"
	"gitlab.com/fynbos/backend/user"
	"gitlab.com/fynbos/backend/wallets"
)

type Backends interface {
	Users() user.Client
	Wallets() wallets.Client
	Gatehub() gatehub.Client
}
