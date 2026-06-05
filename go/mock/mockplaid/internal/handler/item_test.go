package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"gitlab.com/fynbos/mock/mockplaid/internal/models"
	"gitlab.com/fynbos/mock/mockplaid/internal/storage"
)

// seedItem creates a stored item directly and returns it.
func seedItem(t *testing.T, store storage.Storage, instID, name string) models.Item {
	t.Helper()
	item := models.Item{
		AccessToken:     "access-sandbox-x",
		ItemID:          "item-x",
		InstitutionID:   instID,
		InstitutionName: name,
		PublicToken:     "public-sandbox-x",
		Accounts:        []models.Account{{AccountID: "acc_mock_a_checking", Name: "Plaid Checking", Mask: "0000", Type: "depository", Subtype: "checking"}},
	}
	if err := store.SaveItem(context.Background(), item); err != nil {
		t.Fatalf("SaveItem: %v", err)
	}
	return item
}

func post(t *testing.T, fn http.HandlerFunc, body string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	rr := httptest.NewRecorder()
	fn(rr, req)
	return rr
}

func TestExchangePublicToken(t *testing.T) {
	h, store := newTestHandler()
	seedItem(t, store, models.InstitutionA, "Tartan Bank")

	rr := post(t, h.ExchangePublicToken, `{"public_token":"public-sandbox-x"}`)
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
	var resp struct {
		AccessToken string `json:"access_token"`
		ItemID      string `json:"item_id"`
	}
	_ = json.NewDecoder(rr.Body).Decode(&resp)
	if resp.AccessToken != "access-sandbox-x" || resp.ItemID != "item-x" {
		t.Fatalf("unexpected: %+v", resp)
	}
}

func TestExchange_InvalidPublicToken(t *testing.T) {
	h, _ := newTestHandler()
	rr := post(t, h.ExchangePublicToken, `{"public_token":"nope"}`)
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rr.Code)
	}
	var resp struct {
		ErrorCode string `json:"error_code"`
	}
	_ = json.NewDecoder(rr.Body).Decode(&resp)
	if resp.ErrorCode != "INVALID_PUBLIC_TOKEN" {
		t.Fatalf("error_code: %q", resp.ErrorCode)
	}
}

func TestItemGet(t *testing.T) {
	h, store := newTestHandler()
	seedItem(t, store, models.InstitutionA, "Tartan Bank")
	rr := post(t, h.ItemGet, `{"access_token":"access-sandbox-x"}`)
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
	var resp struct {
		Item struct {
			InstitutionID string `json:"institution_id"`
		} `json:"item"`
	}
	_ = json.NewDecoder(rr.Body).Decode(&resp)
	if resp.Item.InstitutionID != models.InstitutionA {
		t.Fatalf("institution_id: %q", resp.Item.InstitutionID)
	}
}

func TestItemGet_InvalidAccessToken(t *testing.T) {
	h, _ := newTestHandler()
	rr := post(t, h.ItemGet, `{"access_token":"nope"}`)
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rr.Code)
	}
}

func TestInstitutionsGetByID(t *testing.T) {
	h, _ := newTestHandler()
	rr := post(t, h.InstitutionsGetByID, `{"institution_id":"ins_mock_a","country_codes":["US"]}`)
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
	var resp struct {
		Institution struct {
			Name string `json:"name"`
		} `json:"institution"`
	}
	_ = json.NewDecoder(rr.Body).Decode(&resp)
	if resp.Institution.Name != "Tartan Bank" {
		t.Fatalf("name: %q", resp.Institution.Name)
	}
}

func TestInstitutionsGetByID_Unknown(t *testing.T) {
	h, _ := newTestHandler()
	rr := post(t, h.InstitutionsGetByID, `{"institution_id":"ins_zzz"}`)
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rr.Code)
	}
}
