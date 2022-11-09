package ops

import (
	"gitlab.com/fynbos/backend/email/sendgrid"
	"gitlab.com/fynbos/backend/kyc"
	"gitlab.com/fynbos/backend/user"
)

type Backends interface {
	External() sendgrid.Client
	Users() user.Client
	KYC() kyc.Client
}
