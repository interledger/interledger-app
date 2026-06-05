package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"testing"

	"gitlab.com/fynbos/mock/mockplaid/internal/models"
	"gitlab.com/fynbos/mock/mockplaid/internal/storage"
)

func TestCreateProcessorToken(t *testing.T) {
	h, store := newTestHandler()
	seedItem(t, store, models.InstitutionA, "Tartan Bank") // account acc_mock_a_checking

	rr := post(t, h.CreateProcessorToken, `{"access_token":"access-sandbox-x","account_id":"acc_mock_a_checking","processor":"fiant"}`)
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d (%s)", rr.Code, rr.Body.String())
	}
	var resp struct {
		ProcessorToken string `json:"processor_token"`
	}
	_ = json.NewDecoder(rr.Body).Decode(&resp)
	if !strings.HasPrefix(resp.ProcessorToken, "processor-sandbox-") {
		t.Fatalf("bad processor_token: %q", resp.ProcessorToken)
	}
}

func TestCreateProcessorToken_BadAccount(t *testing.T) {
	h, store := newTestHandler()
	seedItem(t, store, models.InstitutionA, "Tartan Bank")
	rr := post(t, h.CreateProcessorToken, `{"access_token":"access-sandbox-x","account_id":"acc_zzz","processor":"fiant"}`)
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rr.Code)
	}
}

func TestCreateProcessorToken_BadAccessToken(t *testing.T) {
	h, _ := newTestHandler()
	rr := post(t, h.CreateProcessorToken, `{"access_token":"nope","account_id":"x","processor":"fiant"}`)
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rr.Code)
	}
}

func TestRemoveItem_IdempotentAndDropsItem(t *testing.T) {
	h, store := newTestHandler()
	seedItem(t, store, models.InstitutionB, "Platypus Bank")

	rr := post(t, h.RemoveItem, `{"access_token":"access-sandbox-x"}`)
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
	if _, err := store.GetItemByAccessToken(context.Background(), "access-sandbox-x"); err != storage.ErrItemNotFound {
		t.Fatalf("item not removed: %v", err)
	}
	// second remove is still 200 (idempotent)
	rr2 := post(t, h.RemoveItem, `{"access_token":"access-sandbox-x"}`)
	if rr2.Code != http.StatusOK {
		t.Fatalf("idempotent remove expected 200, got %d", rr2.Code)
	}
}
