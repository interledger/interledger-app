package ops

import (
	"github.com/jmoiron/sqlx"
	"gitlab.com/fynbos/pacioli"
)

type Backends interface {
	Pacioli() pacioli.Client
	LedgerID() uint32
	DB() *sqlx.DB
}
