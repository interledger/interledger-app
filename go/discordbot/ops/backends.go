package ops

import (
	"github.com/jmoiron/sqlx"
	"gitlab.com/fynbos/backend/payments"
)

type Backends interface {
	DB() *sqlx.DB
	Discord() Discord
	Payments() payments.Client
}
