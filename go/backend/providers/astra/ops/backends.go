package ops

import (
	"testing"

	temporal "go.temporal.io/sdk/client"

	"github.com/jmoiron/sqlx"
	"gitlab.com/fynbos/backend/kyc"
	"gitlab.com/fynbos/backend/providers/astra/external"
	"gitlab.com/fynbos/backend/user"
)

type Backends interface {
	DB() *sqlx.DB
	External() external.Client
	Users() user.Client
	KYC() kyc.Client
	Temporal() temporal.Client
}

type ActivityBackends interface {
	DB() *sqlx.DB
	Users() user.Client
	KYC() kyc.Client
	Temporal() temporal.Client
}

type TestBackends struct {
	DBC *sqlx.DB
	//Extr *external_mock.MockClient
}

func (t TestBackends) Temporal() temporal.Client {
	return nil
}

func (t TestBackends) DB() *sqlx.DB {
	return t.DBC
}

func (t TestBackends) External() external.Client {
	return nil //t.Extr
}

func (t TestBackends) Users() user.Client {
	return nil
}

func (t TestBackends) KYC() kyc.Client {
	return nil
}

func NewTestBackends(_ *testing.T, opts ...func(*TestBackends)) Backends {
	b := &TestBackends{}
	for _, opt := range opts {
		opt(b)
	}
	return b
}
