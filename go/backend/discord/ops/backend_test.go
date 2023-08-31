package ops_test

import (
	"github.com/jmoiron/sqlx"
	"gitlab.com/fynbos/backend/discord/external"
	external_mock "gitlab.com/fynbos/backend/discord/external/client/mock"
)

type TestBackends struct {
	Db             *sqlx.DB
	ExternalClient *external_mock.MockClient
}

func (tb *TestBackends) DB() *sqlx.DB {
	return tb.Db
}

func (tb *TestBackends) External() external.Client {
	return tb.ExternalClient
}
