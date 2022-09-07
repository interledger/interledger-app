package client

import (
	"github.com/go-playground/validator/v10"
	"github.com/jmoiron/sqlx"
	"gitlab.com/fynbos/backend/accounts"
	"gitlab.com/fynbos/backend/identity"
	"gitlab.com/fynbos/backend/providers/unit/external"
	"gitlab.com/fynbos/backend/providers/unit/ops"
	temporal_client "go.temporal.io/sdk/client"
)

type Backends interface {
	Validator() *validator.Validate
	DB() *sqlx.DB
	Identity() identity.Client
	Accounts() accounts.Client
	Temporal() temporal_client.Client
}

var _ ops.Backends = opsBackends{}

type opsBackends struct {
	Backends
	unitExternal external.Client
}

func (ob opsBackends) UnitExternal() external.Client {
	return ob.unitExternal
}
