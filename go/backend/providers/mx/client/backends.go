package client

import (
	"github.com/go-playground/validator/v10"
	"github.com/jmoiron/sqlx"
	"gitlab.com/fynbos/backend/accounts"
	"gitlab.com/fynbos/backend/fundingsources"
	"gitlab.com/fynbos/backend/identity"
	"gitlab.com/fynbos/backend/providers/mx"
	"gitlab.com/fynbos/backend/providers/mx/external"
	"gitlab.com/fynbos/backend/providers/mx/ops"
	"gitlab.com/fynbos/backend/providers/unit"
	"gitlab.com/fynbos/backend/twilio"
	temporal "go.temporal.io/sdk/client"
)

type Backends interface {
	Validator() *validator.Validate
	DB() *sqlx.DB
	Accounts() accounts.Client
	Identity() identity.Client
	Temporal() temporal.Client
	Twilio() twilio.Service
	MX() mx.Client
	Unit() unit.Service
	FundingSources() fundingsources.Client
}

var _ ops.Backends = opsBackends{}

type opsBackends struct {
	Backends
	mxExternal external.Mx
}

func (ob opsBackends) MXExternal() external.Mx {
	return ob.mxExternal
}
