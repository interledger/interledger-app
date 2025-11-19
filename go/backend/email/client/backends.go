package client

import (
	kratos "github.com/ory/kratos-client-go"
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
	external sendgrid.Client
	kratos *kratos.APIClient
}

func (o *opsBackends) External() sendgrid.Client {
	return o.external
}

func (o *opsBackends) Kratos() *kratos.APIClient {
	return o.kratos
}
