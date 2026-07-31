package storage

import "testing"

func TestMemoryStore_Contract(t *testing.T) {
	runStoreContractTests(t, func(t *testing.T) Storage {
		return NewMemoryStorage()
	})
}
