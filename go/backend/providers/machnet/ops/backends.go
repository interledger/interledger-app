package ops

import (
	"github.com/jmoiron/sqlx"
	"gitlab.com/fynbos/backend/providers/machnet/external"
)

type Backends interface {
	DB() *sqlx.DB
	External() external.Client
}
