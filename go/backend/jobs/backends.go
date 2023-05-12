package jobs

import (
	"gitlab.com/fynbos/backend/keys"
	"gitlab.com/fynbos/backend/kyc"
	"gitlab.com/fynbos/backend/providers/basistheory"
	"gitlab.com/fynbos/backend/user"
)

type Backends interface {
	Users() user.Client
	Keys() keys.Client
	KYC() kyc.Client
	BasisTheory() basistheory.Client
}

type Activity struct {
	b Backends
}

func NewActivity(b Backends) *Activity {
	return &Activity{b: b}
}
