package ops

import (
	kratos "github.com/ory/kratos-client-go"
	"gitlab.com/fynbos/backend/email/sendgrid"
	"gitlab.com/fynbos/backend/kyc"
	"gitlab.com/fynbos/backend/user"
	"gitlab.com/fynbos/backend/wallets"
)

type Backends interface {
	External() sendgrid.Client
	Users() user.Client
	KYC() kyc.Client
	Wallets() wallets.Client
	Kratos() *kratos.APIClient
}
