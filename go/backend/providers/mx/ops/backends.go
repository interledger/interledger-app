package ops

import (
	"testing"

	"gitlab.com/fynbos/backend/providers/mx/external"
	external_mock "gitlab.com/fynbos/backend/providers/mx/external/client/mock"
)

type Backends interface {
	External() external.Client
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
}

func (b *TestBackends) External() external.Client {
	return b.ExternalClient
}
