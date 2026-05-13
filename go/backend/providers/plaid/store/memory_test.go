package store_test

import (
	"context"
	"sync"
	"testing"
	"time"

	"gitlab.com/fynbos/backend/providers/plaid"
	"gitlab.com/fynbos/backend/providers/plaid/store"
)

func TestMemory_PutGetDelete(t *testing.T) {
	s := store.NewMemory()
	ctx := context.Background()

	if _, ok, err := s.Get(ctx, "user-a"); err != nil || ok {
		t.Fatalf("expected empty store to return ok=false, err=nil; got ok=%v err=%v", ok, err)
	}

	ts := plaid.TokenSet{
		AccessToken:     "access-sandbox-xxx",
		ItemID:          "item-xxx",
		InstitutionID:   "ins_109508",
		InstitutionName: "First Platypus Bank",
		LinkedAt:        time.Now(),
	}
	if err := s.Put(ctx, "user-a", ts); err != nil {
		t.Fatalf("Put: %v", err)
	}

	got, ok, err := s.Get(ctx, "user-a")
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if !ok {
		t.Fatalf("expected ok=true after Put")
	}
	if got.AccessToken != ts.AccessToken || got.ItemID != ts.ItemID {
		t.Fatalf("Get returned mismatched TokenSet: %+v", got)
	}

	if err := s.Delete(ctx, "user-a"); err != nil {
		t.Fatalf("Delete: %v", err)
	}
	if _, ok, _ := s.Get(ctx, "user-a"); ok {
		t.Fatalf("expected ok=false after Delete")
	}
}

func TestMemory_Race(t *testing.T) {
	s := store.NewMemory()
	ctx := context.Background()
	users := []string{"u1", "u2", "u3", "u4"}

	var wg sync.WaitGroup
	for _, u := range users {
		wg.Add(3)
		go func(u string) {
			defer wg.Done()
			for range 1000 {
				_ = s.Put(ctx, u, plaid.TokenSet{AccessToken: u, ItemID: u})
			}
		}(u)
		go func(u string) {
			defer wg.Done()
			for range 1000 {
				_, _, _ = s.Get(ctx, u)
			}
		}(u)
		go func(u string) {
			defer wg.Done()
			for range 1000 {
				_ = s.Delete(ctx, u)
			}
		}(u)
	}
	wg.Wait()
}
