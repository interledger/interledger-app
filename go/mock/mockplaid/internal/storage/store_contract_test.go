package storage

import (
	"context"
	"testing"
	"time"

	"gitlab.com/fynbos/mock/mockplaid/internal/models"
)

type storeFactory func(t *testing.T) Storage

func runStoreContractTests(t *testing.T, newStore storeFactory) {
	t.Helper()

	t.Run("LinkSessionRoundTrip", func(t *testing.T) {
		store := newStore(t)
		s := models.LinkSession{LinkToken: "link-1", UserID: "u_1", CreatedAt: time.Now().UTC()}
		if err := store.SaveLinkSession(context.Background(), s); err != nil {
			t.Fatalf("SaveLinkSession() error: %v", err)
		}
		got, err := store.GetLinkSession(context.Background(), "link-1")
		if err != nil {
			t.Fatalf("GetLinkSession() error: %v", err)
		}
		if got.UserID != "u_1" {
			t.Fatalf("UserID mismatch: got %q want %q", got.UserID, "u_1")
		}
	})

	t.Run("LinkSessionNotFound", func(t *testing.T) {
		store := newStore(t)
		_, err := store.GetLinkSession(context.Background(), "missing")
		if err != ErrLinkSessionNotFound {
			t.Fatalf("error mismatch: got %v want %v", err, ErrLinkSessionNotFound)
		}
	})

	t.Run("ItemResolvableByBothTokens", func(t *testing.T) {
		store := newStore(t)
		item := models.Item{
			AccessToken:     "access-1",
			ItemID:          "item-1",
			InstitutionID:   "ins_mock_a",
			InstitutionName: "Tartan Bank",
			PublicToken:     "public-1",
			Accounts:        []models.Account{{AccountID: "acc_mock_a_checking", Name: "Plaid Checking", Mask: "0000"}},
		}
		if err := store.SaveItem(context.Background(), item); err != nil {
			t.Fatalf("SaveItem() error: %v", err)
		}

		byPub, err := store.GetItemByPublicToken(context.Background(), "public-1")
		if err != nil {
			t.Fatalf("GetItemByPublicToken() error: %v", err)
		}
		if byPub.AccessToken != "access-1" || len(byPub.Accounts) != 1 || byPub.Accounts[0].AccountID != "acc_mock_a_checking" {
			t.Fatalf("byPub mismatch: %+v", byPub)
		}

		byAcc, err := store.GetItemByAccessToken(context.Background(), "access-1")
		if err != nil {
			t.Fatalf("GetItemByAccessToken() error: %v", err)
		}
		if byAcc.ItemID != "item-1" || byAcc.InstitutionName != "Tartan Bank" {
			t.Fatalf("byAcc mismatch: %+v", byAcc)
		}
	})

	t.Run("ItemNotFound", func(t *testing.T) {
		store := newStore(t)
		if _, err := store.GetItemByAccessToken(context.Background(), "nope"); err != ErrItemNotFound {
			t.Fatalf("byAccess error mismatch: got %v want %v", err, ErrItemNotFound)
		}
		if _, err := store.GetItemByPublicToken(context.Background(), "nope"); err != ErrItemNotFound {
			t.Fatalf("byPublic error mismatch: got %v want %v", err, ErrItemNotFound)
		}
	})

	t.Run("DeleteItemIsIdempotentAndDropsBothKeys", func(t *testing.T) {
		store := newStore(t)
		item := models.Item{AccessToken: "access-d", ItemID: "item-d", PublicToken: "public-d"}
		if err := store.SaveItem(context.Background(), item); err != nil {
			t.Fatalf("SaveItem() error: %v", err)
		}
		if err := store.DeleteItemByAccessToken(context.Background(), "access-d"); err != nil {
			t.Fatalf("DeleteItemByAccessToken() error: %v", err)
		}
		if _, err := store.GetItemByAccessToken(context.Background(), "access-d"); err != ErrItemNotFound {
			t.Fatalf("access key not dropped: got %v", err)
		}
		if _, err := store.GetItemByPublicToken(context.Background(), "public-d"); err != ErrItemNotFound {
			t.Fatalf("public key not dropped: got %v", err)
		}
		// second delete is a no-op success
		if err := store.DeleteItemByAccessToken(context.Background(), "access-d"); err != nil {
			t.Fatalf("idempotent delete returned error: %v", err)
		}
	})

	t.Run("NextAccountSeqIsMonotonic", func(t *testing.T) {
		store := newStore(t)
		a, err := store.NextAccountSeq(context.Background())
		if err != nil {
			t.Fatalf("NextAccountSeq() error: %v", err)
		}
		b, err := store.NextAccountSeq(context.Background())
		if err != nil {
			t.Fatalf("NextAccountSeq() error: %v", err)
		}
		if b <= a {
			t.Fatalf("seq not strictly increasing: got %d then %d", a, b)
		}
	})

	t.Run("Reset", func(t *testing.T) {
		store := newStore(t)
		_ = store.SaveLinkSession(context.Background(), models.LinkSession{LinkToken: "r-1"})
		if err := store.Reset(context.Background()); err != nil {
			t.Fatalf("Reset() error: %v", err)
		}
		if _, err := store.GetLinkSession(context.Background(), "r-1"); err != ErrLinkSessionNotFound {
			t.Fatalf("Reset did not clear link sessions: got %v", err)
		}
	})
}
