package client

import (
	"github.com/jmoiron/sqlx"
	"gitlab.com/fynbos/backend/providers/fakecash/ops"
	"gitlab.com/fynbos/pacioli"
)

type Backends interface {
	Pacioli() pacioli.Client
	DB() *sqlx.DB
}

var _ ops.Backends = opsBackends{}

type opsBackends struct {
	ledgerID uint32
	Backends
}

func (b opsBackends) LedgerID() uint32 {
	return b.ledgerID
}
