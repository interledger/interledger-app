package ops

import (
	"gitlab.com/fynbos/backend/email/sendgrid"
	"gitlab.com/fynbos/backend/kyc"
	"gitlab.com/fynbos/backend/user"
	"gitlab.com/fynbos/backend/wallets"
)

type Backends interface {
	External() sendgrid.Client
	OneTemplateID() string
	Users() user.Client
	KYC() kyc.Client
	Wallets() wallets.Client
}
