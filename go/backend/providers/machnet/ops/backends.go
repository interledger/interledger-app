package ops

import (
	"github.com/jmoiron/sqlx"
	"gitlab.com/fynbos/backend/linkedaccounts"
	"gitlab.com/fynbos/backend/kyc"
	"gitlab.com/fynbos/backend/providers/machnet/external"
	"gitlab.com/fynbos/backend/user"
)

type Backends interface {
	DB() *sqlx.DB
	Users() user.Client
	KYC() kyc.Client
	External() external.Client
	LinkedAccounts() linkedaccounts.Client
}
