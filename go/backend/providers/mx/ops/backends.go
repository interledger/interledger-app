package ops

import (
	"testing"

	"gitlab.com/fynbos/backend/kyc"
	kyc_mock "gitlab.com/fynbos/backend/kyc/client/mock"
	"gitlab.com/fynbos/backend/linkedaccounts"
	linkedaccounts_mock "gitlab.com/fynbos/backend/linkedaccounts/client/mock"
	"gitlab.com/fynbos/backend/providers/mx/external"
	external_mock "gitlab.com/fynbos/backend/providers/mx/external/client/mock"
)

type Backends interface {
	External() external.Client
	KYC() kyc.Client
	LinkedAccounts() linkedaccounts.Client
}

var _ Backends = &TestBackends{}

func NewTestBackends(t *testing.T, opts ...func(*TestBackends)) *TestBackends {
	b := &TestBackends{}

	for _, opt := range opts {
		opt(b)
	}

	return b
}

type TestBackends struct {
	ExternalClient *external_mock.MockClient
	Kyc            *kyc_mock.MockClient
	Linkedaccounts *linkedaccounts_mock.MockClient
}

func (b *TestBackends) KYC() kyc.Client {
	return b.Kyc
}

func (b *TestBackends) External() external.Client {
	return b.ExternalClient
}

func (b *TestBackends) LinkedAccounts() linkedaccounts.Client {
	return b.Linkedaccounts
}
