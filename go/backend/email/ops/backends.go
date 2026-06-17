package ops

import (
	"github.com/interledger/interledger-app/go/backend/email/sendgrid"
	"github.com/interledger/interledger-app/go/backend/kyc"
	"github.com/interledger/interledger-app/go/backend/user"
	"github.com/interledger/interledger-app/go/backend/wallets"
)

type Backends interface {
	External() sendgrid.Client
	OneTemplateID() string
	Users() user.Client
	KYC() kyc.Client
	Wallets() wallets.Client
}
