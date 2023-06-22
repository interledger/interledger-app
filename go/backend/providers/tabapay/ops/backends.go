package ops

import (
	"github.com/jmoiron/sqlx"
	"gitlab.com/fynbos/backend/providers/tabapay/external"
	external_mock "gitlab.com/fynbos/backend/providers/tabapay/external/client/mock"
	temporal "go.temporal.io/sdk/client"
)

type Backends interface {
	DB() *sqlx.DB
	External() external.Client
	Temporal() temporal.Client
}

var _ Backends = &TestBackends{}

func NewTestBackends(opts ...func(*TestBackends)) *TestBackends {
	b := &TestBackends{}

	for _, opt := range opts {
		opt(b)
	}

	return b
}

type TestBackends struct {
	Db             *sqlx.DB
	ExternalClient *external_mock.MockClient
	TemporalClient temporal.Client
}

func (b *TestBackends) DB() *sqlx.DB {
	return b.Db
}

func (b *TestBackends) External() external.Client {
	return b.ExternalClient
}

func (b *TestBackends) Temporal() temporal.Client {
	return b.TemporalClient
}
