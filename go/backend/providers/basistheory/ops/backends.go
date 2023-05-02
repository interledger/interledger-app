package ops

import (
	"github.com/jmoiron/sqlx"
	"gitlab.com/fynbos/backend/providers/basistheory/external"
	"gitlab.com/fynbos/backend/providers/basistheory/external/client/mock"
)

type Backends interface {
	DB() *sqlx.DB
	External() external.Client
}

type TestBackends struct {
	Db             *sqlx.DB
	ExternalClient *mock.MockClient
}

func (tb *TestBackends) DB() *sqlx.DB {
	return tb.Db
}

func (tb *TestBackends) External() external.Client {
	return tb.ExternalClient
}
