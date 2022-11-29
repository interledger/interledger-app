package webhook

import (
	"github.com/jmoiron/sqlx"
	"gitlab.com/fynbos/backend/kyc"
	"gitlab.com/fynbos/backend/linkedaccounts"
	"gitlab.com/fynbos/backend/providers/machnet/external"
	"gitlab.com/fynbos/backend/user"
	"go.temporal.io/sdk/client"
)

type Backends interface {
	DB() *sqlx.DB
	Users() user.Client
	KYC() kyc.Client
	LinkedAccounts() linkedaccounts.Client
	Temporal() client.Client
}

type opsBackends struct {
	Backends
	external external.Client
}

func (b opsBackends) External() external.Client {
	return b.external
}
