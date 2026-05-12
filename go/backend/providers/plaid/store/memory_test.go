package store_test

import (
	"sync"
	"testing"
	"time"

	"gitlab.com/fynbos/backend/providers/plaid"
	"gitlab.com/fynbos/backend/providers/plaid/store"
)

func TestMemory_PutGetDelete(t *testing.T) {
	s := store.NewMemory()

	if _, ok := s.Get("user-a"); ok {
		t.Fatalf("expected empty store to return ok=false")
	}

	ts := plaid.TokenSet{
		AccessToken:     "access-sandbox-xxx",
		ItemID:          "item-xxx",
		InstitutionID:   "ins_109508",
		InstitutionName: "First Platypus Bank",
		LinkedAt:        time.Now(),
	}
	s.Put("user-a", ts)

	got, ok := s.Get("user-a")
	if !ok {
		t.Fatalf("expected ok=true after Put")
	}
	if got.AccessToken != ts.AccessToken || got.ItemID != ts.ItemID {
		t.Fatalf("Get returned mismatched TokenSet: %+v", got)
	}

	s.Delete("user-a")
	if _, ok := s.Get("user-a"); ok {
		t.Fatalf("expected ok=false after Delete")
	}
}

// TestMemory_Race exercises concurrent Put/Get/Delete to catch races under
// `go test -race`.
func TestMemory_Race(t *testing.T) {
	s := store.NewMemory()
	users := []string{"u1", "u2", "u3", "u4"}

	var wg sync.WaitGroup
	for _, u := range users {
		wg.Add(3)
		go func(u string) {
			defer wg.Done()
			for range 1000 {
				s.Put(u, plaid.TokenSet{AccessToken: u, ItemID: u})
			}
		}(u)
		go func(u string) {
			defer wg.Done()
			for range 1000 {
				_, _ = s.Get(u)
			}
		}(u)
		go func(u string) {
			defer wg.Done()
			for range 1000 {
				s.Delete(u)
			}
		}(u)
	}
	wg.Wait()
}
