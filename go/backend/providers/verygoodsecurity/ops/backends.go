package ops

import (
	"github.com/jmoiron/sqlx"
	"gitlab.com/fynbos/backend/providers/verygoodsecurity"
)

type Backends interface {
	DB() *sqlx.DB
	VGS() verygoodsecurity.Client
}
