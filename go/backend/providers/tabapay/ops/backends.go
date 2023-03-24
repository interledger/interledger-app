package ops

import (
	"gitlab.com/fynbos/backend/kyc"
	"gitlab.com/fynbos/backend/linkedaccounts"
	"gitlab.com/fynbos/backend/providers/tabapay/external"
)

type Backends interface {
	External() external.Client
	KYC() kyc.Client
	LinkedAccounts() linkedaccounts.Client
}
