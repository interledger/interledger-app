package client

import (
	"github.com/interledger/interledger-app/go/backend/email/sendgrid"
	"github.com/interledger/interledger-app/go/backend/kyc"
	"github.com/interledger/interledger-app/go/backend/user"
	"github.com/interledger/interledger-app/go/backend/wallets"
)

type Backends interface {
	Users() user.Client
	KYC() kyc.Client
	Wallets() wallets.Client
}

type opsBackends struct {
	Backends
	external   sendgrid.Client
	templateID string
}

func (o *opsBackends) External() sendgrid.Client {
	return o.external
}

func (o *opsBackends) OneTemplateID() string {
	return o.templateID
}
