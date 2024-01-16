package ops

import (
	"testing"

	"gitlab.com/fynbos/backend/payments"

	"gitlab.com/fynbos/backend/twilio"

	"github.com/jmoiron/sqlx"
	"gitlab.com/fynbos/backend/kyc"
	"gitlab.com/fynbos/backend/linkedaccounts"
	"gitlab.com/fynbos/backend/providers/astra/external"
	external_mock "gitlab.com/fynbos/backend/providers/astra/external/mock"
	"gitlab.com/fynbos/backend/providers/basistheory"
	"gitlab.com/fynbos/backend/user"
	temporal "go.temporal.io/sdk/client"
)

type Backends interface {
	DB() *sqlx.DB
	External() external.Client
	Users() user.Client
	KYC() kyc.Client
	Temporal() temporal.Client
	BasisTheory() basistheory.Client
	LinkedAccounts() linkedaccounts.Client
	Payments() payments.Client
	Twilio() twilio.Service
}

type ActivityBackends interface {
	DB() *sqlx.DB
	Users() user.Client
	KYC() kyc.Client
	Temporal() temporal.Client
	BasisTheory() basistheory.Client
	LinkedAccounts() linkedaccounts.Client
	Payments() payments.Client
	Twilio() twilio.Service
}

type TestBackends struct {
	DBC  *sqlx.DB
	Extr *external_mock.MockClient
	Ky   kyc.Client
	Uc   user.Client
}

func (t TestBackends) Payments() payments.Client {
	return nil
}

func (t TestBackends) Twilio() twilio.Service {
	return nil
}

func (t TestBackends) BasisTheory() basistheory.Client {
	return nil
}

func (t TestBackends) LinkedAccounts() linkedaccounts.Client {
	return nil
}

func (t TestBackends) Temporal() temporal.Client {
	return nil
}

func (t TestBackends) DB() *sqlx.DB {
	return t.DBC
}

func (t TestBackends) External() external.Client {
	return t.Extr
}

func (t TestBackends) Users() user.Client {
	return t.Uc
}

func (t TestBackends) KYC() kyc.Client {
	return t.Ky
}

func NewTestBackends(_ *testing.T, opts ...func(*TestBackends)) Backends {
	b := &TestBackends{}
	for _, opt := range opts {
		opt(b)
	}
	return b
}
