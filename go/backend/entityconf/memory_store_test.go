package entityconf_test

import (
	"testing"

	"github.com/interledger/interledger-app/go/backend/entityconf"
)

func TestInMemoryStore(t *testing.T) {
	t.Parallel()

	runStoreContractTests(t, func(t *testing.T) entityconf.Store {
		t.Helper()
		return entityconf.NewInMemoryStore()
	})
}
