package store_test

import (
	"context"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"

	"gitlab.com/fynbos/backend/providers/plaid"
	"gitlab.com/fynbos/backend/providers/plaid/store"
)

func newTestRedis(t *testing.T) (*store.Redis, func()) {
	t.Helper()
	mr := miniredis.RunT(t)
	r, err := store.NewRedis("redis://" + mr.Addr() + "/0")
	if err != nil {
		t.Fatalf("NewRedis: %v", err)
	}
	return r, func() {
		_ = r.Close()
		mr.Close()
	}
}

func TestRedis_PutGetDelete(t *testing.T) {
	r, cleanup := newTestRedis(t)
	defer cleanup()
	ctx := context.Background()

	if _, ok, err := r.Get(ctx, "user-a"); err != nil || ok {
		t.Fatalf("expected empty store to return ok=false err=nil; got ok=%v err=%v", ok, err)
	}

	ts := plaid.TokenSet{
		AccessToken:     "access-sandbox-xxx",
		ItemID:          "item-xxx",
		InstitutionID:   "ins_109508",
		InstitutionName: "First Platypus Bank",
		LinkedAt:        time.Now().UTC().Truncate(time.Second),
	}
	if err := r.Put(ctx, "user-a", ts); err != nil {
		t.Fatalf("Put: %v", err)
	}

	got, ok, err := r.Get(ctx, "user-a")
	if err != nil || !ok {
		t.Fatalf("Get: ok=%v err=%v", ok, err)
	}
	if got.AccessToken != ts.AccessToken || got.ItemID != ts.ItemID ||
		got.InstitutionName != ts.InstitutionName || !got.LinkedAt.Equal(ts.LinkedAt) {
		t.Fatalf("Get returned mismatched TokenSet: %+v vs %+v", got, ts)
	}

	if err := r.Delete(ctx, "user-a"); err != nil {
		t.Fatalf("Delete: %v", err)
	}
	if _, ok, _ := r.Get(ctx, "user-a"); ok {
		t.Fatalf("expected ok=false after Delete")
	}
}

func TestRedis_BadURL(t *testing.T) {
	if _, err := store.NewRedis("not-a-url"); err == nil {
		t.Fatalf("expected error for invalid URL")
	}
}
