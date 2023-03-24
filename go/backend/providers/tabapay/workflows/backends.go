package workflows

import (
	"gitlab.com/fynbos/backend/kyc"
	kyc_mock "gitlab.com/fynbos/backend/kyc/client/mock"
	"gitlab.com/fynbos/backend/linkedaccounts"
	linkedaccount_mock "gitlab.com/fynbos/backend/linkedaccounts/client/mock"
	"gitlab.com/fynbos/backend/providers/tabapay/external"
	mock_client "gitlab.com/fynbos/backend/providers/tabapay/external/client/mock"
)

type Backends interface {
	External() external.Client
	KYC() kyc.Client
	LinkedAccounts() linkedaccounts.Client
}

type TestBackends struct {
	ExternalClient *mock_client.MockClient
	Kyc            *kyc_mock.MockClient
	Linkedaccounts *linkedaccount_mock.MockClient
}

func (tb *TestBackends) External() external.Client {
	return tb.ExternalClient
}

func (tb *TestBackends) KYC() kyc.Client {
	return tb.Kyc
}

func (tb *TestBackends) LinkedAccounts() linkedaccounts.Client {
	return tb.Linkedaccounts
}

func NewTestBackends(opts ...func(b *TestBackends)) *TestBackends {
	b := &TestBackends{}

	for _, opt := range opts {
		opt(b)
	}

	return b
}
