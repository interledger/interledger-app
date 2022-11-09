package client

import (
	"gitlab.com/fynbos/backend/email/sendgrid"
	"gitlab.com/fynbos/backend/kyc"
	"gitlab.com/fynbos/backend/user"
)

type Backends interface {
	Users() user.Client
	KYC() kyc.Client
}

type opsBackends struct {
	Backends
	external sendgrid.Client
}

func (o *opsBackends) External() sendgrid.Client {
	return o.external
}
