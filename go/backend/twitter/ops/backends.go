package ops

import (
	"github.com/jmoiron/sqlx"
	"gitlab.com/fynbos/backend/twitter/external"
	external_mock "gitlab.com/fynbos/backend/twitter/external/client/mock"
	temporal "go.temporal.io/sdk/client"
)

type (
	Backends interface {
		DB() *sqlx.DB
		External() external.Client
		Temporal() temporal.Client
	}

	TestBackends struct {
		Db             *sqlx.DB
		ExternalClient *external_mock.MockClient
	}
)

func (tb *TestBackends) Temporal() temporal.Client {
	//TODO implement me
	panic("implement me")
}

func (tb *TestBackends) DB() *sqlx.DB {
	return tb.Db
}

func (tb *TestBackends) External() external.Client {
	return tb.ExternalClient
}

func NewTestBackends(opts ...func(*TestBackends)) *TestBackends {
	b := &TestBackends{}

	for _, opt := range opts {
		opt(b)
	}

	return b
}
