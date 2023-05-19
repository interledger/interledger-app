package jobs

import (
	"github.com/jmoiron/sqlx"
	"gitlab.com/fynbos/backend/keys"
	"gitlab.com/fynbos/backend/kyc"
	"gitlab.com/fynbos/backend/linkedaccounts"
	"gitlab.com/fynbos/backend/providers/basistheory"
	"gitlab.com/fynbos/backend/user"
)

type Backends interface {
	Users() user.Client
	Keys() keys.Client
	KYC() kyc.Client
	DB() *sqlx.DB
	LinkedAccounts() linkedaccounts.Client
	BasisTheory() basistheory.Client
}

type Activity struct {
	b Backends
}

func NewActivity(b Backends) *Activity {
	return &Activity{b: b}
}
