package storage

import (
	"testing"

	"github.com/alicebob/miniredis/v2"
)

func TestRedisStore_Contract(t *testing.T) {
	runStoreContractTests(t, func(t *testing.T) Storage {
		mini, err := miniredis.Run()
		if err != nil {
			t.Fatalf("miniredis.Run() unexpected error: %v", err)
		}
		t.Cleanup(mini.Close)

		store, err := NewRedisStorage("redis://"+mini.Addr(), 0)
		if err != nil {
			t.Fatalf("NewRedisStorage() unexpected error: %v", err)
		}
		return store
	})
}
