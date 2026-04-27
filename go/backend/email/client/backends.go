package client

import (
	"gitlab.com/fynbos/backend/email/sendgrid"
	"gitlab.com/fynbos/backend/kyc"
	"gitlab.com/fynbos/backend/user"
	"gitlab.com/fynbos/backend/wallets"
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
