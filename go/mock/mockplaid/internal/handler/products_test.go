package handler

import (
	"encoding/json"
	"net/http"
	"testing"

	"gitlab.com/fynbos/mock/mockplaid/internal/models"
)

func TestGetAccounts(t *testing.T) {
	h, store := newTestHandler()
	seedItem(t, store, models.InstitutionA, "Tartan Bank")
	rr := post(t, h.GetAccounts, `{"access_token":"access-sandbox-x"}`)
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
	var resp struct {
		Accounts []struct {
			AccountID string `json:"account_id"`
			Balances  struct {
				Current float64 `json:"current"`
			} `json:"balances"`
		} `json:"accounts"`
	}
	_ = json.NewDecoder(rr.Body).Decode(&resp)
	if len(resp.Accounts) != 1 || resp.Accounts[0].AccountID != "acc_mock_a_checking" {
		t.Fatalf("accounts mismatch: %+v", resp.Accounts)
	}
	if resp.Accounts[0].Balances.Current == 0 {
		t.Fatal("missing balance")
	}
}

func TestGetAuth(t *testing.T) {
	h, store := newTestHandler()
	seedItem(t, store, models.InstitutionA, "Tartan Bank")
	rr := post(t, h.GetAuth, `{"access_token":"access-sandbox-x"}`)
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
	var resp struct {
		Numbers struct {
			ACH []struct {
				AccountID string `json:"account_id"`
				Routing   string `json:"routing"`
			} `json:"ach"`
		} `json:"numbers"`
	}
	_ = json.NewDecoder(rr.Body).Decode(&resp)
	if len(resp.Numbers.ACH) != 1 || resp.Numbers.ACH[0].AccountID != "acc_mock_a_checking" || resp.Numbers.ACH[0].Routing == "" {
		t.Fatalf("ach mismatch: %+v", resp.Numbers.ACH)
	}
}

func TestGetIdentity(t *testing.T) {
	h, store := newTestHandler()
	seedItem(t, store, models.InstitutionA, "Tartan Bank")
	rr := post(t, h.GetIdentity, `{"access_token":"access-sandbox-x"}`)
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
	var resp struct {
		Accounts []struct {
			Owners []struct {
				Names []string `json:"names"`
			} `json:"owners"`
		} `json:"accounts"`
	}
	_ = json.NewDecoder(rr.Body).Decode(&resp)
	if len(resp.Accounts) != 1 || len(resp.Accounts[0].Owners) != 1 || resp.Accounts[0].Owners[0].Names[0] != "Alex Mock" {
		t.Fatalf("identity mismatch: %+v", resp.Accounts)
	}
}

func TestTransactionsSync(t *testing.T) {
	h, store := newTestHandler()
	seedItem(t, store, models.InstitutionA, "Tartan Bank")
	rr := post(t, h.TransactionsSync, `{"access_token":"access-sandbox-x"}`)
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
	var resp struct {
		Added      []map[string]interface{} `json:"added"`
		HasMore    bool                     `json:"has_more"`
		NextCursor string                   `json:"next_cursor"`
	}
	_ = json.NewDecoder(rr.Body).Decode(&resp)
	if resp.HasMore {
		t.Fatal("has_more must be false (terminal page)")
	}
	if len(resp.Added) != 1 || resp.NextCursor != "mockplaid-cursor-end" {
		t.Fatalf("sync mismatch: added=%d cursor=%q", len(resp.Added), resp.NextCursor)
	}
}

func TestProducts_InvalidAccessToken(t *testing.T) {
	h, _ := newTestHandler()
	for _, fn := range []http.HandlerFunc{h.GetAccounts, h.GetAuth, h.GetBalance, h.GetIdentity, h.TransactionsSync} {
		rr := post(t, fn, `{"access_token":"nope"}`)
		if rr.Code != http.StatusBadRequest {
			t.Fatalf("expected 400 for invalid token, got %d", rr.Code)
		}
	}
}
