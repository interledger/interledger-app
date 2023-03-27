package ops

import (
	"gitlab.com/fynbos/backend/kyc"
	"gitlab.com/fynbos/backend/linkedaccounts"
	"gitlab.com/fynbos/backend/providers/tabapay/external"
	temporal "go.temporal.io/sdk/client"
)

type Backends interface {
	External() external.Client
	KYC() kyc.Client
	LinkedAccounts() linkedaccounts.Client
	Temporal() temporal.Client
}
