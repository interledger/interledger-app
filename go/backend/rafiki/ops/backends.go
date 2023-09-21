package ops

import (
	"github.com/jmoiron/sqlx"
	"gitlab.com/fynbos/backend/payments"
	"gitlab.com/fynbos/backend/rafiki/external"
)

type Backends interface {
	DB() *sqlx.DB
	External() external.Client
	Payments() payments.Client
}

type ActivityBackends interface {
	DB() *sqlx.DB
	Payments() payments.Client
}
